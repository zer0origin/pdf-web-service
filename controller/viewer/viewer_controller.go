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

	token, err := t.KeycloakApi.AuthenticateJwtToken(c.GetString(keycloak.AccessTokenKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ownerUid, err := token.Claims.GetSubject()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	meta, err := t.JesrApi.GetMeta(documentUid, ownerUid)
	if err != nil {
		if errors.Is(err, jesr.GetMetaNotFoundError) {
			fmt.Println("User meta data not found. Creating now!")
			_ = NotificationService.GetServiceInstance().SendMessage(ownerUid, "User meta data not found! Attempting to create that now.") //TODO Error method!

			err := t.JesrApi.AddMeta(jesr.AddMetaRequest{
				DocumentUUID: uuid.MustParse(documentUid),
				OwnerUUID:    uuid.MustParse(ownerUid),
			})
			if err != nil {
				fmt.Println("Failed to create user meta")
				_ = NotificationService.GetServiceInstance().SendMessage(ownerUid, "Failed to create user meta.") //TODO Error method!
			} else {
				_ = NotificationService.GetServiceInstance().SendMessage(ownerUid, "Successfully created user meta.") //TODO Error method!

				t.GetViewer(c)
				return
			}
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve document meta"})
		return
	}

	c.HTML(http.StatusOK, "viewer", meta)
}
