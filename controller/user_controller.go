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
	"strconv"
)

type UserController struct {
	KeycloakApi *keycloak.Api
	JesrApi     jesr.Api
}

func (t UserController) UserInfo(c *gin.Context) {
	accessToken := c.GetString(keycloak.AccessTokenKey)
	user, err := t.KeycloakApi.SendUserInfoRequest(accessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	accept := c.Request.Header["Content-Type"][0]
	if accept == "application/x-www-form-urlencoded" {
		baseString, baseStringPresent := c.GetPostForm("documentBase64String")
		documentTile, documentTilePresent := c.GetPostForm("documentTitle")
		ownerTypeString, ownerTypeStringPresent := c.GetPostForm("ownerType")

		if !baseStringPresent || !documentTilePresent || !ownerTypeStringPresent {
			c.JSON(http.StatusBadRequest, gin.H{"error": errors.New("missing required form values")})
			return
		}

		ownerType, err := strconv.Atoi(ownerTypeString)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		data := jesr.UploadRequest{
			DocumentBase64String: baseString,
			OwnerType:            ownerType,
			DocumentTitle:        documentTile,
		}

		token, err := t.KeycloakApi.AuthenticateJwtToken(c.GetString(keycloak.AccessTokenKey))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		subject, err := token.Claims.GetSubject()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if data.OwnerType != 1 {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": errors.New("owner type unsupported")})
			return
		}

		data.OwnerUUID = uuid.MustParse(subject)
		err = t.JesrApi.UploadDocument(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		documentsOwnerByUser, err := t.JesrApi.GetDocuments(uuid.MustParse(subject))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.HTML(http.StatusOK, "userdata", documentsOwnerByUser)
		c.Status(http.StatusCreated)
		return
	}

	c.JSON(http.StatusBadRequest, "Unsupported accept header")
}
