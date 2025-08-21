package user

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"pdf_service_web/jesr"
	"pdf_service_web/keycloak"
	"pdf_service_web/models"
	"pdf_service_web/service/NotificationService"
	"strconv"
)

type GinUser struct {
	KeycloakApi *keycloak.Api
	JesrApi     jesr.Api
}

func (t GinUser) AppBase(c *gin.Context) {
	data := models.PageDefaults{
		NavDetails: &models.NavDetails{IsAuthenticated: true},
	}
	c.HTML(http.StatusOK, "base", data)
}

func (t GinUser) UserInfo(c *gin.Context) {
	accessToken := c.GetString(keycloak.AccessTokenKey)
	user, err := t.KeycloakApi.SendUserInfoRequest(accessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	data := models.PageDefaults{
		NavDetails:           &models.NavDetails{IsAuthenticated: true},
		ContentDetails:       user,
		NotificationSettings: &models.NotificationSettings{Uid: user.Uid},
	}
	c.HTML(http.StatusOK, "userinfo", data)
}

type ContentDetails struct {
	PageInfo models.PageInfo
	UserData any
}

func (t GinUser) UserDashboard(c *gin.Context) {
	token, err := t.KeycloakApi.AuthenticateJwtToken(c.GetString(keycloak.AccessTokenKey))
	if err != nil {
		fmt.Println(fmt.Errorf("failed to authenticate access token in user dashboard %s", err.Error()))
		c.SetCookie(keycloak.AccessTokenKey, "", -1, "", "", false, false)
		c.SetCookie(keycloak.RefreshTokenKey, "", -1, "", "", false, false)
		c.Redirect(http.StatusTemporaryRedirect, "/") //Login page
		return
	}

	var limit int8 = 15
	if limitValue, present := c.GetQuery("limit"); present {
		parseInt, err := strconv.ParseInt(limitValue, 10, 8)
		if err != nil {
			return
		}

		limit = int8(parseInt)
	}

	var offset int8 = 0
	if offsetValue, present := c.GetQuery("offset"); present {
		parseInt, err := strconv.ParseInt(offsetValue, 10, 8)
		if err != nil {
			return
		}

		offset = int8(parseInt)
	}

	if offset < 0 {
		offset = 0
	}

	if limit <= 1 {
		limit = 1
	}

	subject, _ := token.Claims.GetSubject()
	documentsOwnerByUser, _ := t.JesrApi.GetDocumentsByOwnerUUID(uuid.MustParse(subject), limit, offset)

	if offset != 0 && len(documentsOwnerByUser) == 0 {
		offset = 0
		documentsOwnerByUser, _ = t.JesrApi.GetDocumentsByOwnerUUID(uuid.MustParse(subject), limit, offset)
	}

	data := models.PageDefaults{
		NavDetails: &models.NavDetails{IsAuthenticated: true},
		ContentDetails: ContentDetails{
			PageInfo: models.PageInfo{
				Offset:   int(offset),
				NextPage: int(offset + limit),
				LastPage: int(offset - limit),
				Limit:    int(limit),
			},
			UserData: documentsOwnerByUser,
		},
		NotificationSettings: &models.NotificationSettings{Uid: subject},
	}
	c.HTML(http.StatusOK, "userdashboard", data)
}

func (t GinUser) Upload(c *gin.Context) {
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
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

		results := make(chan error)
		go func() {
			results <- t.JesrApi.UploadDocument(data)
			//SEND FOR RESULTS FOR PROCESSING
		}()

		select {
		case err = <-results:
			instance := NotificationService.GetServiceInstance()
			if err != nil {
				_ = instance.SendEvent(subject, "DocumentDelete", err.Error())
			}

			_ = instance.SendEvent(subject, "DocumentDelete", "Success")
		}
	}

	c.JSON(http.StatusBadRequest, "Unsupported accept header")
}

func (t GinUser) DeleteDocument(c *gin.Context) {
	documentUidStr := c.Param("uid")
	uid, err := uuid.Parse(documentUidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse uid"})
		return
	}
	token, err := t.KeycloakApi.AuthenticateJwtToken(c.GetString(keycloak.AccessTokenKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ownerUuidStr, err := token.Claims.GetSubject()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resultsChannel := make(chan error)
	go func() {
		resultsChannel <- t.JesrApi.DeleteDocuments(uid, uuid.MustParse(ownerUuidStr))
	}()

	select {
	case err = <-resultsChannel:
		instance := NotificationService.GetServiceInstance()
		if err != nil {
			_ = instance.SendEvent(ownerUuidStr, "DocumentDelete", err.Error())
		}

		_ = instance.SendEvent(ownerUuidStr, "DocumentDelete", "Success")
	}

	if err != nil {
		fmt.Println(err.Error())
	}
}

func (t GinUser) PushNotifications(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Writer.Flush()

	cookie, err := c.Request.Cookie(keycloak.AccessTokenKey)
	if err != nil {
		_, _ = fmt.Fprint(c.Writer, fmt.Sprintf("data: %s\n\n", "<script>window.location.href = \"/\"</script>"))
		return
	}

	token, err := t.KeycloakApi.AuthenticateJwtToken(cookie.Value)
	if err != nil {
		_, _ = fmt.Fprint(c.Writer, fmt.Sprintf("data: %s\n\n", "<script>window.location.href = \"/\"</script>"))
		return
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return
	}

	notificationService := NotificationService.GetServiceInstance()
	notificationChannel := notificationService.GetOrCreateChannel(subject)
	defer notificationService.DeleteNotificationChannel(subject)

	clientGone := c.Request.Context().Done()
	for {
		select {
		case msg := <-notificationChannel.Channel:
			_, err := fmt.Fprint(c.Writer, msg)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			c.Writer.Flush()
		case <-clientGone:
			fmt.Println("Client has disconnected!")

			return
		}
	}
}

type Broadcast struct {
	Message string `json:"message"`
}

func (t GinUser) BroadcastNotification(c *gin.Context) {
	bc := &Broadcast{}
	err := c.ShouldBindJSON(bc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	notificationService := NotificationService.GetServiceInstance()
	notificationService.Broadcast(bc.Message)

	c.Status(http.StatusOK)
}
