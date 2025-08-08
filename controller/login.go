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
	if _, err := c.Request.Cookie("refreshToken"); err == nil {

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

		authAttempt, statusCode, err := SendLoginAuthAttemptWithPasswordAndUsername(username, password)
		if err != nil {
			return
		}

		if statusCode != http.StatusOK {
			fmt.Printf("Returned Server Status Code: \n%d\n", statusCode)
			c.JSON(statusCode, gin.H{})
			return
		}

		res := &model.TokenResponse{}
		err = json.Unmarshal([]byte(authAttempt), &res)
		if err != nil {
			return
		}

		c.SetCookie("accessToken", res.AccessToken, res.AccessExpiresIn, "", "", false, false)
		c.SetCookie("refreshToken", res.RefreshToken, res.RefreshExpiresIn, "", "", false, false)
		c.SetCookie("idToken", res.IdToken, res.RefreshExpiresIn, "", "", false, false)

		c.Header("HX-Redirect", "/User/UserInfo")
		c.Status(http.StatusOK)
		return
	}

	if accept == "application/json" || accept == "*/*" {
		loginInfo := &model.UnauthenticatedUser{}
		err := c.ShouldBindJSON(loginInfo)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}

		authAttempt, statusCode, err := SendLoginAuthAttemptWithPasswordAndUsername(loginInfo.Username, loginInfo.Password)
		if statusCode != http.StatusOK {
			fmt.Printf("Returned Server Status Code: \n%d\n", statusCode)
			c.JSON(statusCode, gin.H{})
			return
		}

		var dat map[string]interface{}
		err = json.Unmarshal([]byte(authAttempt), &dat)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}

		c.JSON(http.StatusOK, dat)
		return
	}

	c.JSON(http.StatusUnprocessableEntity, gin.H{})
}

func SendLoginAuthAttemptWithPasswordAndUsername(username, password string) (string, int, error) {
	url := "http://localhost:8081/realms/pdf/protocol/openid-connect/token"
	method := "POST"

	payload := strings.NewReader("grant_type=password&audience=service-api&username=" + username + "&password=" + password + "&scope=openid%20profile%20email%20organization")

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return "", -1, fmt.Errorf("error creating new request: %s", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	credentials := fmt.Sprintf("%s:%s", "service-api", "gtQLem8EJgxr537nbQlJh3Npd6Li6s0K")
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	authHeaderValue := "Basic " + encodedCredentials
	req.Header.Add("Authorization", authHeaderValue)

	res, err := client.Do(req)
	if err != nil {
		return "", -1, fmt.Errorf("error sending http request: %s", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", res.StatusCode, fmt.Errorf("error reading http response: %s", err)
	}

	fmt.Println(string(body))
	return string(body), res.StatusCode, nil
}

func GetAccessToken() {

}
