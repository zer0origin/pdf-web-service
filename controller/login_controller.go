package controller

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"pdf_service_web/model"
	"strings"
)

type LoginController struct {
}

func (LoginController) LoginRender(c *gin.Context) {
	if refreshTokenCookie, err := c.Request.Cookie("refreshToken"); err == nil {
		if _, err := c.Request.Cookie("accessToken"); err == nil {
			c.Redirect(http.StatusFound, "/user/info")
			return
		}

		token, err := SendLoginAuthAttemptWithRefreshToken(refreshTokenCookie.Value)
		if err == nil {
			c.SetCookie("accessToken", token.AccessToken, token.AccessExpiresIn, "/", "", false, false)
			c.SetCookie("refreshToken", token.RefreshToken, token.RefreshExpiresIn, "/", "", false, false)
			c.SetCookie("idToken", token.IdToken, token.RefreshExpiresIn, "/", "", false, false)
			c.Redirect(http.StatusFound, "/user/info")
			return
		}
	}

	c.HTML(http.StatusOK, "login", gin.H{})
}

func (LoginController) LoginAuthHandler(c *gin.Context) {
	accept := c.Request.Header["Accept"][0]
	if accept == "text/html" {
		username, userPresent := c.GetPostForm("username")
		password, passPresent := c.GetPostForm("password")

		if !userPresent || !passPresent {
			c.JSON(http.StatusUnprocessableEntity, gin.H{})
			return
		}

		authUser, err := SendLoginAuthAttemptWithPasswordAndUsername(username, password)
		if err != nil {
			c.JSON(http.StatusBadRequest, err) //TODO RETURN ERROR AS HTML
			return
		}

		c.SetCookie("accessToken", authUser.AccessToken, authUser.AccessExpiresIn, "/", "", false, false)
		c.SetCookie("refreshToken", authUser.RefreshToken, authUser.RefreshExpiresIn, "/", "", false, false)
		c.SetCookie("idToken", authUser.IdToken, authUser.RefreshExpiresIn, "/", "", false, false)
		c.Header("HX-Redirect", "/user/info")
		c.Status(http.StatusOK)
		return
	}

	if accept == "application/json" || accept == "*/*" {
		loginInfo := &model.UnauthenticatedUser{}
		err := c.ShouldBindJSON(loginInfo)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		authAttempt, err := SendLoginAuthAttemptWithPasswordAndUsername(loginInfo.Username, loginInfo.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		c.JSON(http.StatusOK, authAttempt)
		return
	}

	c.JSON(http.StatusUnprocessableEntity, gin.H{})
}

func SendLoginAuthAttemptWithPasswordAndUsername(username, password string) (model.TokenResponse, error) {
	url := "http://localhost:8081/realms/pdf/protocol/openid-connect/token"
	method := "POST"

	payload := strings.NewReader("grant_type=password&audience=service-api&username=" + username + "&password=" + password + "&scope=openid%20profile%20email%20organization")

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return model.TokenResponse{}, fmt.Errorf("error creating new request: %s", err)
	}

	request, err := handleLoginAuthRequest(req)
	return request, err
}

func SendLoginAuthAttemptWithRefreshToken(refreshToken string) (model.TokenResponse, error) {
	url := "http://localhost:8081/realms/pdf/protocol/openid-connect/token"
	method := "POST"

	payload := strings.NewReader("grant_type=refresh_token&audience=service-api&refresh_token=" + refreshToken + "&scope=openid%20profile%20email%20organization")

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return model.TokenResponse{}, fmt.Errorf("error creating new request: %s", err)
	}

	request, err := handleLoginAuthRequest(req)
	return request, nil
}

func handleLoginAuthRequest(req *http.Request) (model.TokenResponse, error) {
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	credentials := fmt.Sprintf("%s:%s", "service-api", "gtQLem8EJgxr537nbQlJh3Npd6Li6s0K")
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	authHeaderValue := "Basic " + encodedCredentials
	req.Header.Add("Authorization", authHeaderValue)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return model.TokenResponse{}, fmt.Errorf("error sending http request: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return model.TokenResponse{}, fmt.Errorf("request unauthorized")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return model.TokenResponse{}, fmt.Errorf("error reading http response: %s", err)
	}

	fmt.Println()

	token := &model.TokenResponse{}
	err = json.Unmarshal(body, &token)
	return *token, nil
}
