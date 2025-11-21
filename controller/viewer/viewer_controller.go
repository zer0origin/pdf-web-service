package viewer

import (
	"errors"
	"fmt"
	"net/http"
	"pdf_service_web/jesr"
	"pdf_service_web/keycloak"
	"pdf_service_web/service/NotificationService"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GinViewer struct {
	KeycloakApi *keycloak.Api
	JesrApi     jesr.Api
}

func (t GinViewer) GetViewer(c *gin.Context) {
	documentUid := c.Param("uid")
	_, err := uuid.Parse(documentUid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse uid"})
		return
	}

	token, err := t.KeycloakApi.ParseTokenUnverified(c.GetString(keycloak.AccessTokenKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ownerUid, err := token.Claims.GetSubject()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cookieStr, err := c.Cookie("client_id")
	if err != nil {
		fmt.Println("Cookie for user " + ownerUid + " not found")
		return
	}

	meta, err := t.JesrApi.GetMeta(documentUid, ownerUid)
	if err != nil {
		if errors.Is(err, jesr.GetMetaNotFoundError) {
			fmt.Println("User meta data not found. Creating now!")
			_ = NotificationService.GetServiceInstance().SendMessage(cookieStr, "User meta data not found! Attempting to create that now.")

			err := t.JesrApi.AddMeta(jesr.AddMetaRequest{
				DocumentUUID: uuid.MustParse(documentUid),
				OwnerUUID:    uuid.MustParse(ownerUid),
			})
			if err != nil {
				fmt.Println("Failed to create user meta")
				_ = NotificationService.GetServiceInstance().SendMessage(cookieStr, "Failed to create user meta.")
			} else {
				_ = NotificationService.GetServiceInstance().SendMessage(cookieStr, "Successfully created user meta.")

				t.GetViewer(c)
				return
			}
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve document meta"})
		return
	}

	c.HTML(http.StatusOK, "viewer", meta)
}
