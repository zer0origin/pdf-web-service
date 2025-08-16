package controller

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"pdf_service_web/controller/models"
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
	data := models.PageDefaults{
		NavDetails:     models.NavDetails{IsAuthenticated: true},
		ContentDetails: user,
	}
	c.HTML(http.StatusOK, "userinfo", data)
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

	data := models.PageDefaults{
		NavDetails:     models.NavDetails{IsAuthenticated: true},
		ContentDetails: documentsOwnerByUser,
	}
	c.HTML(http.StatusOK, "userdashboard", data)
}

func (t UserController) Upload(c *gin.Context) {
	accept := c.Request.Header["Accept"][0]
	if accept == "application/json" || accept == "*/*" {
		data := &jesr.UploadRequest{}

		err := c.ShouldBindJSON(data)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		token, err := t.KeycloakApi.AuthenticateJwtToken(c.GetString(keycloak.AccessTokenKey))
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		subject, err := token.Claims.GetSubject()
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		if data.OwnerType != 1 {
			c.JSON(http.StatusUnprocessableEntity, errors.New("owner type unsupported"))
			return
		}

		data.OwnerUUID = uuid.MustParse(subject)
		err = t.JesrApi.UploadDocument(*data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusCreated)
		return
	}
	c.JSON(http.StatusBadRequest, "Unsupported accept header")
}
