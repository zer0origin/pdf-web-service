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
	"pdf_service_web/service/NotificationService"
	"strconv"
	"time"
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
		NavDetails:           models.NavDetails{IsAuthenticated: true},
		ContentDetails:       user,
		NotificationSettings: &models.NotificationSettings{Uid: user.Uid},
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
		NavDetails:           models.NavDetails{IsAuthenticated: true},
		ContentDetails:       documentsOwnerByUser,
		NotificationSettings: &models.NotificationSettings{Uid: subject},
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

func (t UserController) PushNotifications(c *gin.Context) {
	fmt.Println("Request Received!")

	cookie, err := c.Request.Cookie(keycloak.AccessTokenKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	token, err := t.KeycloakApi.AuthenticateJwtToken(cookie.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return
	}

	uid := c.Param("uid")
	if uid != subject {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errors.New("you are not authorized to read " + uid + " event stream")})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	// Important: Flush the headers to the client immediately
	c.Writer.Flush()

	notificationService := NotificationService.GetInstance()
	userChannel := notificationService.CreateNotificationChannel(uid)
	defer notificationService.DeleteNotificationChannel(uid)

	clientGone := c.Request.Context().Done()
	for {
		select {
		case msg := <-userChannel:
			bytesWritten, err := fmt.Fprint(c.Writer, msg)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			c.Writer.Flush()
			fmt.Printf("Bytes sent to client: %d\n", bytesWritten) // Added newline for cleaner output
		case <-clientGone:
			fmt.Println("Client has disconnected!")

			return
		}
	}
}

func (t UserController) BroadcastNotification(c *gin.Context) {
	notificationService := NotificationService.GetInstance()
	notificationService.Broadcast(fmt.Sprintf("data: <div>HELLO CONNECTED CLIENT %d</div>\n\n", time.Now().Unix()))

	c.Status(http.StatusOK)
}
