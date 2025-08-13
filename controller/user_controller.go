package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"pdf_service_web/jesr"
	"pdf_service_web/keycloak"
)

type UserController struct {
	KeycloakApi *keycloak.Api
	JesrApi     jesr.Api
}

func (t UserController) UserInfo(c *gin.Context) {
	accessToken := c.GetString(keycloak.AccessTokenKey)
	user, err := t.KeycloakApi.SendUserInfoRequest(accessToken)
	if err != nil {
		return
	}

	c.HTML(http.StatusOK, "userinfo", user)
}

func (t UserController) UserDashboard(c *gin.Context) {
	token, err := t.KeycloakApi.AuthenticateJwtToken(c.GetString(keycloak.AccessTokenKey))
	if err != nil {
		fmt.Println(fmt.Errorf("failed to authenticate access token in user dashboard %s", err.Error()))
		c.SetCookie(keycloak.AccessTokenKey, "", -1, "", "", false, false)
		c.SetCookie(keycloak.RefreshTokenKey, "", -1, "", "", false, false)
		c.Redirect(http.StatusTemporaryRedirect, "/") //Login page
		return
	}

	subject, _ := token.Claims.GetSubject()
	documentsOwnerByUser, _ := t.JesrApi.GetDocuments(uuid.MustParse(subject))
	c.HTML(http.StatusOK, "userdashboard", documentsOwnerByUser)
}
