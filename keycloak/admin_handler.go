package keycloak

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type AdminHandler struct {
	LoginContext *context.Context
	realmConfig  RealmConfig
}

type NewUser struct {
	Username      string           `json:"username"`
	Email         string           `json:"email"`
	Enabled       bool             `json:"enabled"`
	VerifiedEmail bool             `json:"emailVerified"`
	Credentials   []NewCredentials `json:"credentials"`
}

type NewCredentials struct {
	PasswordType string `json:"type"`
	Value        string `json:"value"`
	Temporary    bool   `json:"temporary"`
}

func (t *AdminHandler) Token() (*TokenResponse, error) {
	loginContext := *t.LoginContext
	if loginContext == nil || loginContext.Err() != nil || loginContext.Value("JwtToken") == nil {
		ctx, err := getAdminLoginToken(t.realmConfig)
		if err != nil {
			return &TokenResponse{}, err
		}

		loginContext = ctx
	}

	response, ok := loginContext.Value("JwtToken").(*TokenResponse)
	if !ok {
		return &TokenResponse{}, errors.New("even after token refresh, 'JwtToken' not found in context")
	}

	return response, nil
}

func NewAdminHandler(realmConfig RealmConfig) (AdminHandler, error) {
	ctx, err := getAdminLoginToken(realmConfig)
	if err != nil {
		return AdminHandler{}, err
	}

	return AdminHandler{
		LoginContext: &ctx,
		realmConfig:  realmConfig,
	}, nil
}

func getAdminLoginToken(realmConfig RealmConfig) (context.Context, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", realmConfig.BaseUrl, realmConfig.RealmName)
	method := "POST"
	payload := strings.NewReader("grant_type=client_credentials&client_id=" + realmConfig.Client + "&client_secret=" + realmConfig.ClientSecret)
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	token := &TokenResponse{}
	err = json.Unmarshal(body, token)
	if err != nil {
		return nil, err
	}

	ctxTimeout, _ := context.WithTimeout(context.Background(), time.Duration(token.AccessExpiresIn)*time.Second)
	ctxValue := context.WithValue(ctxTimeout, "JwtToken", token)
	fmt.Println("Refreshed service-api admin account access token.")
	return ctxValue, nil
}

type FailedToCreateUserDueToConflictError struct {
	ErrorMessage string `json:"errorMessage"`
}

func (t *AdminHandler) CreateNewUserWithPassword(username, email, password string, enabled, verifiedEmail bool) error {
	url := fmt.Sprintf("%s/admin/realms/%s/users", t.realmConfig.BaseUrl, t.realmConfig.RealmName)
	method := "POST"

	newUser := NewUser{
		Username:      username,
		Email:         email,
		Enabled:       enabled,
		VerifiedEmail: verifiedEmail,
		Credentials: []NewCredentials{{
			PasswordType: "password",
			Value:        password,
			Temporary:    false,
		}},
	}

	bytes, err := json.Marshal(newUser)
	if err != nil {
		return err
	}

	payload := strings.NewReader(string(bytes))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	token, err := t.Token()
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusConflict {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		errorToReturn := FailedToCreateUserDueToConflictError{}
		err = json.Unmarshal(body, &errorToReturn)
		if err != nil {
			return err
		}
		return errors.New(errorToReturn.ErrorMessage)
	}

	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("server encountered an unexpected error %d", res.StatusCode)
	}

	return nil
}
