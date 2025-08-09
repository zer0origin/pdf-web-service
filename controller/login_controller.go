package controller

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"pdf_service_web/keycloak"
	"strings"
)

type LoginController struct {
	AuthenticatedRedirect string
	RealmConfig           keycloak.RealmConfig
}

func (t LoginController) LoginRender(c *gin.Context) {
	if refreshTokenCookie, err := c.Request.Cookie("refreshToken"); err == nil {
		if _, err := c.Request.Cookie("accessToken"); err == nil {
			c.Redirect(http.StatusFound, t.AuthenticatedRedirect)
			return
		}

		token, err := t.sendLoginAuthAttemptWithRefreshToken(refreshTokenCookie.Value)
		if err == nil {
			c.SetCookie("accessToken", token.AccessToken, token.AccessExpiresIn, "/", "", false, false)
			c.SetCookie("refreshToken", token.RefreshToken, token.RefreshExpiresIn, "/", "", false, false)
			c.SetCookie("idToken", token.IdToken, token.RefreshExpiresIn, "/", "", false, false)
			c.Redirect(http.StatusFound, t.AuthenticatedRedirect)
			return
		}
	}

	c.HTML(http.StatusOK, "login", gin.H{})
}

func (t LoginController) LoginAuthHandler(c *gin.Context) {
	accept := c.Request.Header["Accept"][0]
	if accept == "text/html" {
		username, userPresent := c.GetPostForm("username")
		password, passPresent := c.GetPostForm("password")

		if !userPresent || !passPresent {
			c.JSON(http.StatusUnprocessableEntity, gin.H{})
			return
		}

		authUser, err := t.sendLoginAuthAttemptWithPasswordAndUsername(username, password)
		if err != nil {
			c.JSON(http.StatusBadRequest, err) //TODO RETURN ERROR AS HTML
			return
		}

		c.SetCookie("accessToken", authUser.AccessToken, authUser.AccessExpiresIn, "/", "", false, false)
		c.SetCookie("refreshToken", authUser.RefreshToken, authUser.RefreshExpiresIn, "/", "", false, false)
		c.SetCookie("idToken", authUser.IdToken, authUser.RefreshExpiresIn, "/", "", false, false)
		c.Header("HX-Redirect", t.AuthenticatedRedirect)
		c.Status(http.StatusOK)
		return
	}

	if accept == "application/json" || accept == "*/*" {
		loginInfo := &keycloak.UnauthenticatedUser{}
		err := c.ShouldBindJSON(loginInfo)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		authAttempt, err := t.sendLoginAuthAttemptWithPasswordAndUsername(loginInfo.Username, loginInfo.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		c.JSON(http.StatusOK, authAttempt)
		return
	}

	c.JSON(http.StatusUnprocessableEntity, gin.H{})
}

func (t LoginController) sendLoginAuthAttemptWithPasswordAndUsername(username, password string) (keycloak.TokenResponse, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", t.RealmConfig.BaseUrl, t.RealmConfig.RealmName)
	method := "POST"

	payload := strings.NewReader("grant_type=password&audience=" + t.RealmConfig.Client + "&username=" + username + "&password=" + password + "&scope=openid%20profile%20email%20organization")

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return keycloak.TokenResponse{}, fmt.Errorf("error creating new request: %s", err)
	}

	request, err := t.handleLoginAuthRequest(req)
	return request, err
}

func (t LoginController) sendLoginAuthAttemptWithRefreshToken(refreshToken string) (keycloak.TokenResponse, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", t.RealmConfig.BaseUrl, t.RealmConfig.RealmName)
	method := "POST"

	payload := strings.NewReader("grant_type=refresh_token&audience=" + t.RealmConfig.Client + "&refresh_token=" + refreshToken + "&scope=openid%20profile%20email%20organization")

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return keycloak.TokenResponse{}, fmt.Errorf("error creating new request: %s", err)
	}

	request, err := t.handleLoginAuthRequest(req)
	return request, nil
}

func (t LoginController) handleLoginAuthRequest(req *http.Request) (keycloak.TokenResponse, error) {
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	credentials := fmt.Sprintf("%s:%s", t.RealmConfig.Client, t.RealmConfig.ClientSecret)
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	authHeaderValue := "Basic " + encodedCredentials
	req.Header.Add("Authorization", authHeaderValue)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return keycloak.TokenResponse{}, fmt.Errorf("error sending http request: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return keycloak.TokenResponse{}, fmt.Errorf("request unauthorized")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return keycloak.TokenResponse{}, fmt.Errorf("error reading http response: %s", err)
	}

	token := &keycloak.TokenResponse{}
	err = json.Unmarshal(body, &token)
	return *token, nil
}
