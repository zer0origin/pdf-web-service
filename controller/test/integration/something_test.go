package integration

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

var dbUser = "user"
var dbPassword = "password"
var dbDatabase = "postgres"

func TestFunc(t *testing.T) {
	serviceHostname, servicePort := setupApiService(t)

	apiBaseUrl := fmt.Sprintf("http://%s:%s", serviceHostname, servicePort)
	post, err := http.Get(fmt.Sprintf("%s/ping", apiBaseUrl))
	require.NoError(t, err)

	fmt.Printf("%s", post.Status)
}

func setupApiService(t *testing.T) (string, string) {
	networkContext := context.Background()
	newNetwork, err := network.New(networkContext)
	testcontainers.CleanupNetwork(t, newNetwork)
	require.NoError(t, err)
	networkName := newNetwork.Name

	wd, err := os.Getwd()
	require.NoError(t, err)
	sqlScript := wd + "/test-container/sql/BasicSetup.sql"

	postgresContext := context.Background()
	postgresContainer, err := postgres.Run(
		postgresContext,
		"postgres:16-alpine",
		postgres.WithInitScripts(sqlScript),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		network.WithNetwork([]string{"postgres"}, newNetwork),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)

	req := testcontainers.ContainerRequest{
		Image:          "pdf_service_api:1.0.2",
		Networks:       []string{networkName},
		NetworkAliases: map[string][]string{networkName: {"api_service"}},
		ExposedPorts:   []string{"8080/tcp"},
		WaitingFor:     wait.ForLog("[GIN-debug] Listening and serving HTTP on :8080").WithOccurrence(1).WithStartupTimeout(5 * time.Second),
		Env: map[string]string{
			"DATABASE_USER":     dbUser,
			"DATABASE_PASSWORD": dbPassword,
			"DATABASE_PORT":     "5432",
			"DATABASE_HOST":     "postgres",
			"DATABASE_DB":       dbDatabase},
	}

	serviceContext := context.Background()
	serviceContainer, err := testcontainers.GenericContainer(serviceContext, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer func() {
		if err := serviceContainer.Terminate(serviceContext); err != nil {
			log.Fatalf("Failed to terminate container: %v", err)
		}

		if err := postgresContainer.Terminate(postgresContext); err != nil {
			log.Fatalf("Failed to terminate container: %v", err)
		}
	}()

	port, err := serviceContainer.MappedPort(serviceContext, "8080")
	require.NoError(t, err)
	fmt.Printf("api container listening to: %s\n", port.Port())

	host, err := serviceContainer.Host(serviceContext)
	require.NoError(t, err)

	return host, port.Port()
}
