package viewer

import (
	"errors"
	"fmt"
	"net/http"
	"pdf_service_web/jesr"
	"pdf_service_web/keycloak"

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
			err := t.JesrApi.AddMeta(jesr.AddMetaRequest{
				DocumentUUID: uuid.MustParse(documentUid),
				OwnerUUID:    uuid.MustParse(ownerUid),
			})
			if err != nil {
				fmt.Println("Failed to create user meta")
			} else {
				fmt.Println("Successfully created user meta")
				t.GetViewer(c)
				return
			}
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve document meta"})
		return
	}

	c.HTML(http.StatusOK, "viewer", meta)
}
