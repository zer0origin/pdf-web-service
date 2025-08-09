package keycloak

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type RealmConfig struct {
	BaseUrl      string
	RealmName    string
	Client       string
	ClientSecret string
}

func (t RealmConfig) SendLoginAuthAttemptWithPasswordAndUsername(username, password string) (TokenResponse, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", t.BaseUrl, t.RealmName)
	method := "POST"

	payload := strings.NewReader("grant_type=password&audience=" + t.Client + "&username=" + username + "&password=" + password + "&scope=openid%20profile%20email%20organization")

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("error creating new request: %s", err)
	}

	request, err := t.handleLoginAuthRequest(req)
	return request, err
}

func (t RealmConfig) SendLoginAuthAttemptWithRefreshToken(refreshToken string) (TokenResponse, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", t.BaseUrl, t.RealmName)
	method := "POST"

	payload := strings.NewReader("grant_type=refresh_token&audience=" + t.Client + "&refresh_token=" + refreshToken + "&scope=openid%20profile%20email%20organization")

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("error creating new request: %s", err)
	}

	request, err := t.handleLoginAuthRequest(req)
	return request, nil
}

func (t RealmConfig) handleLoginAuthRequest(req *http.Request) (TokenResponse, error) {
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	credentials := fmt.Sprintf("%s:%s", t.Client, t.ClientSecret)
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	authHeaderValue := "Basic " + encodedCredentials
	req.Header.Add("Authorization", authHeaderValue)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("error sending http request: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return TokenResponse{}, fmt.Errorf("request unauthorized")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("error reading http response: %s", err)
	}

	token := &TokenResponse{}
	err = json.Unmarshal(body, &token)
	return *token, nil
}
