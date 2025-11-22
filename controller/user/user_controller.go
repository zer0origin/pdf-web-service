package user

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"pdf_service_web/jesr"
	"pdf_service_web/keycloak"
	"pdf_service_web/models"
	"pdf_service_web/service/NotificationService"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	token, err := t.KeycloakApi.ParseTokenUnverified(c.GetString(keycloak.AccessTokenKey))
	if err != nil {
		fmt.Println(fmt.Errorf("failed to parse access token in user dashboard %s", err.Error()))
		c.SetCookie(keycloak.AccessTokenKey, "", -1, "", "", false, false)
		c.SetCookie(keycloak.RefreshTokenKey, "", -1, "", "", false, false)
		c.Redirect(http.StatusTemporaryRedirect, "/") //Login page
		return
	}

	var limit uint32 = 15
	if limitValue, present := c.GetQuery("limit"); present {
		parseInt, err := strconv.ParseInt(limitValue, 10, 8)
		if err != nil {
			return
		}

		limit = uint32(parseInt)
	}

	var offset uint32 = 0
	if offsetValue, present := c.GetQuery("offset"); present {
		parseInt, err := strconv.ParseInt(offsetValue, 10, 8)
		if err != nil {
			return
		}

		offset = uint32(parseInt)
	}

	if offset < 0 {
		offset = 0
	}

	if limit <= 1 {
		limit = 15
	}

	subject, _ := token.Claims.GetSubject()
	documentsOwnerByUser, err := t.JesrApi.GetDocumentsByOwnerUUID(uuid.MustParse(subject), limit, offset)
	if err != nil {
		log.Printf("failed to connect to database: %v", err.Error())
	}

	if offset != 0 && len(documentsOwnerByUser) == 0 {
		offset = 0
		documentsOwnerByUser, err = t.JesrApi.GetDocumentsByOwnerUUID(uuid.MustParse(subject), limit, offset)
		if err != nil {
			log.Printf("failed to connect to database: %v", err.Error())
		}
	}

	nextPage := offset + limit
	lastPage := offset - limit
	data := models.PageDefaults{
		NavDetails: &models.NavDetails{IsAuthenticated: true},
		ContentDetails: ContentDetails{
			PageInfo: models.PageInfo{
				Offset:   offset,
				NextPage: &nextPage,
				LastPage: &lastPage,
				Limit:    limit,
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

		uploadRequest := jesr.UploadRequest{
			DocumentBase64String: baseString,
			OwnerType:            ownerType,
			DocumentTitle:        documentTile,
		}

		token, err := t.KeycloakApi.ParseTokenUnverified(c.GetString(keycloak.AccessTokenKey))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		subject, err := token.Claims.GetSubject()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if uploadRequest.OwnerType != 1 {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": errors.New("owner type unsupported")})
			return
		}

		uploadRequest.OwnerUUID = uuid.MustParse(subject)
		cookieStr, err := c.Cookie("client_id")
		if err != nil {
			fmt.Println("Cookie for user " + subject + " not found")
			c.Redirect(http.StatusFound, "/")
			return
		}

		instance := NotificationService.GetServiceInstance()

		results := make(chan error)
		go func() {
			_ = instance.SendMessage(cookieStr, "Uploading document!")
			docUUID, err := t.JesrApi.UploadDocument(uploadRequest)
			if err != nil {
				results <- err
				return
			}

			metaRequest := jesr.AddMetaRequest{
				DocumentUUID:         docUUID,
				OwnerUUID:            uploadRequest.OwnerUUID,
				OwnerType:            uploadRequest.OwnerType,
				DocumentBase64String: &uploadRequest.DocumentBase64String,
			}

			results <- t.JesrApi.AddMeta(metaRequest)
		}()

		select {
		case err = <-results:
			if err != nil {
				fmt.Printf("Error uploading %s document: %s\n", subject, err.Error())
				_ = instance.SendMessage(cookieStr, "Error uploading document!")
				return
			}

			instance.SendEventToAllInstancesOfUser(cookieStr, "DocumentUpload", "Success")
		}

		return
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
	token, err := t.KeycloakApi.ParseTokenUnverified(c.GetString(keycloak.AccessTokenKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ownerUuidStr, err := token.Claims.GetSubject()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cookieStr, err := c.Cookie("client_id")
	if err != nil {
		fmt.Println("Cookie for user " + ownerUuidStr + " not found")
		c.Redirect(http.StatusFound, "/")
		return
	}

	resultsChannel := make(chan error)
	go func() {
		_ = NotificationService.GetServiceInstance().SendMessage(cookieStr, "Deleting document")
		resultsChannel <- t.JesrApi.DeleteDocuments(uid, uuid.MustParse(ownerUuidStr))
	}()

	select {
	case err = <-resultsChannel:
		instance := NotificationService.GetServiceInstance()
		if err != nil {
			_ = NotificationService.GetServiceInstance().SendEvent(cookieStr, "errorNotif", "Failed to delete document!")
		}

		instance.SendEventToAllInstancesOfUser(cookieStr, "DocumentDelete", "Success")
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

	cookie, err := c.Request.Cookie(keycloak.AccessTokenKey)
	if err != nil {
		_, _ = fmt.Fprint(c.Writer, fmt.Sprintf("event: refresh\ndata: %s\n\n", "Token Rejected"))
		return
	}

	token, err := t.KeycloakApi.ParseTokenUnverified(cookie.Value)
	if err != nil {
		_, _ = fmt.Fprint(c.Writer, fmt.Sprintf("event: refresh\ndata: %s\n\n", "Token Rejected"))
		return
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return
	}
	cookieStr, err := c.Cookie("client_id")
	if cookieStr == "" {
		fmt.Println("Creating new client_id")
		nUUID := uuid.NewString()
		cookieStr = fmt.Sprintf("%s.%s", subject, nUUID[len(nUUID)-12:])
		c.SetCookie("client_id", cookieStr, 60*60*60*24, "/", "", false, false)

		_, err := fmt.Fprint(c.Writer, "")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		c.Writer.Flush()
	}

	notificationService := NotificationService.GetServiceInstance()
	notificationChannel, err := notificationService.GetOrCreateNotificationChannel(cookieStr)
	defer notificationService.DeleteNotificationChannel(cookieStr)
	if err != nil {
		fmt.Println("Failed to create notification channel for " + subject + " using " + cookieStr)
		return
	}

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
