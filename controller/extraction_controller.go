package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"pdf_service_web/jesr"
	"pdf_service_web/keycloak"
	"pdf_service_web/models"
)

type GinExtraction struct {
	KeycloakApi *keycloak.Api
	JesrApi     jesr.Api
}

func (t GinExtraction) BasicExtraction(c *gin.Context) {
	cookie, err := c.Request.Cookie(keycloak.AccessTokenKey)
	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}

	token, err := t.KeycloakApi.ParseTokenUnverified(cookie.Value)
	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}

	res, err := t.JesrApi.ExtractSelection(c, subject)
	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}

	c.HTML(http.StatusOK, "extractionContent", res)
}

func ConvertExtractionResponseCSVToTable(resp models.ExtractionResponse) []string {
	return nil
}
