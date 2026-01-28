package viewer

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

type GinViewer struct {
	KeycloakApi *keycloak.Api
	JesrApi     jesr.Api
}

type ContentDetails struct {
	PageInfo models.PageInfo
	PageData any
}

type ViewerData struct {
	ViewData    models.Meta
	DocumentUid string
}

func (t GinViewer) GetViewer(c *gin.Context) {
	documentUid := c.Param("uid")
	_, err := uuid.Parse(documentUid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse uid"})
		return
	}

	c.HTML(http.StatusOK, "viewer", gin.H{"DocumentUid": documentUid})
}

func (t GinViewer) GetImages(c *gin.Context) {
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

	var offset uint32 = 0
	if offsetValue, present := c.GetQuery("offset"); present {
		parseInt, err := strconv.ParseInt(offsetValue, 10, 8)
		if err != nil {
			return
		}

		offset = uint32(parseInt)
	}

	var limit uint32 = 5
	if limitValue, present := c.GetQuery("limit"); present {
		parseInt, err := strconv.ParseInt(limitValue, 10, 8)
		if err != nil {
			return
		}

		limit = uint32(parseInt)
	}

	if limit <= 0 {
		limit = 5
	}

	if offset < 0 {
		offset = 0
	}

	meta, err := t.JesrApi.GetMeta(documentUid, ownerUid, offset, limit)
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

		fmt.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve document meta"})
		return
	}

	nextPage := offset + limit
	lastPage := offset - limit

	newModel := models.PageDefaults{
		ContentDetails: ContentDetails{
			PageInfo: models.PageInfo{
				Offset: offset,
				NextPage: func() *uint32 {
					if nextPage >= uint32(*meta.NumberOfPages) {
						return nil
					}
					return &nextPage
				}(),
				LastPage: func() *uint32 {
					if lastPage >= nextPage {
						return nil
					}
					return &lastPage
				}(),
				Limit: limit,
			},
			PageData: ViewerData{
				ViewData:    meta,
				DocumentUid: documentUid,
			},
		},
	}

	c.HTML(http.StatusOK, "viewerImages", newModel)
}

func (t GinViewer) UploadSelections(c *gin.Context) {
	//TODO: Check that the user has access to this document, before adding the selections.
	str, err := t.JesrApi.AddSelectionsBulk(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save selection data"})
		return
	}

	c.JSON(200, str)
}

func (t GinViewer) LoadSelections(c *gin.Context) {
	//TODO: Check that the user has access to this document, before adding the selections.
	str, err := t.JesrApi.GetSelectionListString(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load selection data"})
		return
	}

	c.JSON(200, str)
}
