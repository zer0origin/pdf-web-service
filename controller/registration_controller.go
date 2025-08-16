package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/mail"
	"pdf_service_web/controller/models"
	"pdf_service_web/keycloak"
)

type RegistrationController struct {
	CreatedUserRedirect string
	KeycloakApi         *keycloak.Api
}

func (t RegistrationController) RegisterHandle(c *gin.Context) {
	accept := c.Request.Header["Accept"][0]
	if accept == "text/html" || accept == "*/*" {
		username, _ := c.GetPostForm("username")
		email, _ := c.GetPostForm("email")
		password, _ := c.GetPostForm("password")

		if username == "" || password == "" || email == "" {
			errorToSend := models.BasicError{ErrorMessage: "Fill in all text boxes!"}
			c.HTML(http.StatusUnprocessableEntity, "errorMessage", errorToSend)
			return
		}

		if !validEmail(email) {
			errorToSend := models.BasicError{ErrorMessage: "Invalid email address!"}
			c.HTML(http.StatusUnprocessableEntity, "errorMessage", errorToSend)
			return
		}

		fmt.Printf("Registration: %s, %s, %s\n", username, email, password)
		err := t.KeycloakApi.CreateNewUserWithPassword(username, email, password, true, false)

		if err != nil {
			fmt.Println(err.Error())
			errorToSend := models.BasicError{ErrorMessage: err.Error()}
			c.HTML(http.StatusUnprocessableEntity, "errorMessage", errorToSend)
			return
		}

		c.Header("HX-Redirect", t.CreatedUserRedirect)
		c.Status(http.StatusOK)
		return
	}

	c.JSON(http.StatusBadRequest, "Unsupported accept header")
}

func (t RegistrationController) RegisterRender(c *gin.Context) {
	c.HTML(http.StatusOK, "register", models.PageDefaults{
		NavDetails:     models.NavDetails{},
		ContentDetails: gin.H{},
	})
}

func validEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
