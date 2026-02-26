package user

import (
	"errors"
	"fmt"
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
		_ = c.Error(err)
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
		_ = c.Error(err)
		c.SetCookie(keycloak.AccessTokenKey, "", -1, "", "", false, false)
		c.SetCookie(keycloak.RefreshTokenKey, "", -1, "", "", false, false)
		c.Redirect(http.StatusTemporaryRedirect, "/") //Login page
		return
	}

	var limit uint32 = 15
	if limitValue, present := c.GetQuery("limit"); present {
		parseInt, err := strconv.ParseInt(limitValue, 10, 8)
		if err != nil {
			_ = c.Error(err)
			return
		}

		limit = uint32(parseInt)
	}

	var offset uint32 = 0
	if offsetValue, present := c.GetQuery("offset"); present {
		parseInt, err := strconv.ParseInt(offsetValue, 10, 8)
		if err != nil {
			_ = c.Error(err)
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
		_ = c.Error(err)
		return
	}

	if offset != 0 && len(documentsOwnerByUser) == 0 {
		offset = 0
		documentsOwnerByUser, err = t.JesrApi.GetDocumentsByOwnerUUID(uuid.MustParse(subject), limit, offset)
		if err != nil {
			_ = c.Error(err)
			return
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

type UploadRequest struct {
	BaseString      string `json:"documentBase64String"`
	DocumentTile    string `json:"documentTitle"`
	OwnerTypeString string `json:"ownerType"`
}

func (t GinUser) Upload(c *gin.Context) {
	accept := c.Request.Header["Content-Type"][0]
	var baseString, documentTile, ownerTypeString string
	var baseStringPresent, documentTilePresent, ownerTypeStringPresent bool

	if accept == "application/x-www-form-urlencoded" {
		baseString, baseStringPresent = c.GetPostForm("documentBase64String")
		documentTile, documentTilePresent = c.GetPostForm("documentTitle")
		ownerTypeString, ownerTypeStringPresent = c.GetPostForm("ownerType")
	}

	if accept == "application/json" {
		req := &UploadRequest{}
		err := c.ShouldBindBodyWithJSON(req)
		if err != nil {
			_ = c.Error(errors.New("failed to parse body"))
			return
		}

		baseString = req.BaseString
		documentTile = req.DocumentTile
		ownerTypeString = req.OwnerTypeString

		baseStringPresent = req.BaseString != ""
		documentTilePresent = req.DocumentTile != ""
		ownerTypeStringPresent = req.OwnerTypeString != ""
	}

	if !baseStringPresent || !documentTilePresent || !ownerTypeStringPresent {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required form values"})
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
		_ = c.Error(err)
		return
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		_ = c.Error(err)
		return
	}

	if uploadRequest.OwnerType != 1 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "owner type unsupported"})
		return
	}

	uploadRequest.OwnerUUID = uuid.MustParse(subject)
	clientUid, err := c.Cookie("client_id")
	if err != nil {
		fmt.Println("client_id for user " + subject + " not found")
	}

	instance := NotificationService.GetServiceInstance()

	_ = instance.SendMessage(clientUid, "Uploading document!")
	docUUID, err := t.JesrApi.UploadDocument(uploadRequest)
	if err != nil {
		fmt.Printf("Error uploading %s document: %s\n", subject, err.Error())
		_ = instance.SendMessage(clientUid, "Error uploading document!")
		_ = c.Error(errors.New("error uploading document"))
		return
	}

	metaRequest := jesr.AddMetaRequest{
		DocumentUUID:         docUUID,
		OwnerUUID:            uploadRequest.OwnerUUID,
		OwnerType:            uploadRequest.OwnerType,
		DocumentBase64String: &uploadRequest.DocumentBase64String,
	}

	err = t.JesrApi.AddMeta(metaRequest)
	if err != nil {
		fmt.Printf("Error uploading %s document: %s\n", subject, err.Error())
		err = instance.SendMessage(clientUid, "Error uploading document!")
		_ = c.Error(errors.New("error uploading document"))
		return
	}

	if clientUid == "" {
		c.Status(http.StatusFound)
		return
	}
	_ = instance.SendEvent(clientUid, "DocumentUpload", "Success")
	c.Status(http.StatusOK)
	return

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
		_ = c.Error(err)
		return
	}

	ownerUuidStr, err := token.Claims.GetSubject()
	if err != nil {
		_ = c.Error(err)
		return
	}

	clientUid, err := c.Cookie("client_id")
	if err != nil {
		fmt.Println("Cookie for user " + ownerUuidStr + " not found")
	}

	instance := NotificationService.GetServiceInstance()
	_ = instance.SendMessage(clientUid, "Deleting document")
	err = t.JesrApi.DeleteDocuments(uid, uuid.MustParse(ownerUuidStr))

	if err != nil {
		_ = instance.SendEvent(clientUid, "errorNotif", "Failed to delete document!")
		_ = c.Error(err)
		return
	}

	if clientUid == "" {
		c.Status(http.StatusFound)
		return
	}

	_ = instance.SendEvent(clientUid, "DocumentDelete", "Success")
	c.Status(http.StatusOK)
}

func (t GinUser) ToastNotifications(c *gin.Context) {
	c.Status(http.StatusBadGateway)
	if true {
		return
	}

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
		_, _ = fmt.Fprint(c.Writer, fmt.Sprintf("event: refresh\ndata: %s\n\n", "Token Rejected"))
		return
	}

	cookieStr, err := c.Cookie("client_id")
	if cookieStr == "" {
		fmt.Println("Creating new client_id")
		nUUID := uuid.NewString()
		cookieStr = fmt.Sprintf("%s.%s", subject, nUUID[len(nUUID)-12:])
		c.SetCookie("client_id", cookieStr, -1, "/", "", false, false)

		_, err := fmt.Fprint(c.Writer, "")
		c.Writer.Flush()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
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
