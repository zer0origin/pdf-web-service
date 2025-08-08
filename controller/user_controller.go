package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"pdf_service_web/model"
)

type UserController struct {
}

func (UserController) UserInfo(c *gin.Context) {
	accessTokenCookie, err := c.Request.Cookie("accessToken")
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	user, err := sendUserInfoRequest(accessTokenCookie.Value)
	if err != nil {
		return
	}

	c.HTML(http.StatusOK, "userinfo", user)
}

func sendUserInfoRequest(accessToken string) (model.AuthenticatedUser, error) {
	url := "http://localhost:8081/realms/pdf/protocol/openid-connect/userinfo"
	method := "GET"
	authHeaderValue := fmt.Sprintf("Bearer %s", accessToken)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return model.AuthenticatedUser{}, err
	}

	req.Header.Add("Authorization", authHeaderValue)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return model.AuthenticatedUser{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	user := &model.AuthenticatedUser{}
	err = json.Unmarshal(body, user)
	return *user, err
}
