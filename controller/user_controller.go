package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pdf_service_web/keycloak"
)

type UserController struct {
	Keycloak keycloak.Keycloak
}

func (t UserController) UserInfo(c *gin.Context) {
	accessToken := c.GetString(keycloak.AccessTokenKey)
	user, err := t.Keycloak.SendUserInfoRequest(accessToken)
	if err != nil {
		return
	}

	c.HTML(http.StatusOK, "userinfo", user)
}
