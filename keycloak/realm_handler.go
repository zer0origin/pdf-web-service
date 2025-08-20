package keycloak

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

type RealmHandler struct {
	BaseUrl      string
	RealmName    string
	Client       string
	ClientSecret string
	PublicKey    string
	mutex        sync.Mutex
}

// SendUserInfoRequest
// Send a request to the API server, grab the users details from the response and parse the data to keycloak.AuthenticatedUser
func (t *RealmHandler) SendUserInfoRequest(accessToken string) (AuthenticatedUser, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/userinfo", t.BaseUrl, t.RealmName)
	method := "GET"
	authHeaderValue := fmt.Sprintf("Bearer %s", accessToken)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return AuthenticatedUser{}, err
	}

	req.Header.Add("Authorization", authHeaderValue)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return AuthenticatedUser{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	user := &AuthenticatedUser{}
	err = json.Unmarshal(body, user)
	return *user, err
}

func (t *RealmHandler) SendLoginAuthAttemptWithPasswordAndUsername(username, password string) (TokenResponse, error) {
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

func (t *RealmHandler) SendLoginAuthAttemptWithRefreshToken(refreshToken string) (TokenResponse, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", t.BaseUrl, t.RealmName)
	method := "POST"

	payload := strings.NewReader("grant_type=refresh_token&audience=" + t.Client + "&refresh_token=" + refreshToken + "&scope=openid%20profile%20email%20organization")

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("error creating new request: %s", err)
	}

	request, err := t.handleLoginAuthRequest(req)
	return request, err
}

func (t *RealmHandler) handleLoginAuthRequest(req *http.Request) (TokenResponse, error) {
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

func (t *RealmHandler) GetSigningKey() (string, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.PublicKey == "" {
		fmt.Println("[keycloak] Contacting keycloak, obtaining realm public key")
		url := fmt.Sprintf("%s/realms/%s/", t.BaseUrl, t.RealmName)
		method := "GET"

		client := &http.Client{}
		req, err := http.NewRequest(method, url, nil)

		if err != nil {
			fmt.Println(err)
			return "", err
		}
		res, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return "", err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		type KeyRequest struct {
			PublicKey string `json:"public_key"`
		}
		data := &KeyRequest{}
		err = json.Unmarshal(body, data)
		if err != nil {
			return "", err
		}

		t.PublicKey = data.PublicKey
	}

	return fmt.Sprintf(`-----BEGIN PUBLIC KEY-----
%s
-----END PUBLIC KEY-----`, t.PublicKey), nil
}
