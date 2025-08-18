package login

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"pdf_service_web/controller"
	"pdf_service_web/keycloak"
	models2 "pdf_service_web/models"
)

type GinLogin struct {
	AuthenticatedRedirect string
	Keycloak              *keycloak.Api
	Middleware            controller.GinMiddleware
}

var onSucceeded = func(c *gin.Context, accessToken string) {
	c.Set(keycloak.AccessTokenKey, accessToken)
}

func (t GinLogin) BaseRender(c *gin.Context) {
	redirect := func(accessToken string) {
		c.Redirect(http.StatusFound, t.AuthenticatedRedirect)
	}

	signin := func() {
		c.HTML(http.StatusOK, "base", models2.PageDefaults{
			ContentDetails: gin.H{},
		})
	}

	if err := t.Middleware.IsAuthenticated(c, false, signin, redirect); err != nil {
		fmt.Println(err)
	}
}

func (t GinLogin) LoginRender(c *gin.Context) {
	accept := c.Request.Header["Accept"][0]
	if accept == "text/html" {
		c.HTML(http.StatusOK, "login", models2.PageDefaults{
			ContentDetails: gin.H{},
		})

		return
	}

	c.JSON(http.StatusBadRequest, "Unsupported accept header")
}

func (t GinLogin) LoginAuthHandler(c *gin.Context) {
	accept := c.Request.Header["Accept"][0]
	if accept == "text/html" {
		username, userPresent := c.GetPostForm("username")
		password, passPresent := c.GetPostForm("password")

		if !userPresent || !passPresent || username == "" || password == "" {
			errorToSend := models2.BasicError{ErrorMessage: "Fill in all text boxes!"}
			c.HTML(http.StatusUnprocessableEntity, "errorMessage", errorToSend)
			return
		}

		authUser, err := t.Keycloak.SendLoginAuthAttemptWithPasswordAndUsername(username, password)
		if err != nil {
			errorToSend := models2.BasicError{ErrorMessage: "Incorrect username or password"}
			c.HTML(http.StatusUnauthorized, "errorMessage", errorToSend)
			return
		}

		c.SetCookie("accessToken", authUser.AccessToken, authUser.AccessExpiresIn, "/", "", false, false)
		c.SetCookie("refreshToken", authUser.RefreshToken, authUser.RefreshExpiresIn, "/", "", false, false)
		c.SetCookie("idToken", authUser.IdToken, authUser.RefreshExpiresIn, "/", "", false, false)
		c.Header("HX-Redirect", t.AuthenticatedRedirect)
		return
	}

	if accept == "application/json" || accept == "*/*" {
		loginInfo := &keycloak.UnauthenticatedUser{}
		err := c.ShouldBindJSON(loginInfo)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		authAttempt, err := t.Keycloak.SendLoginAuthAttemptWithPasswordAndUsername(loginInfo.Username, loginInfo.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		c.JSON(http.StatusOK, authAttempt)
		return
	}

	c.JSON(http.StatusBadRequest, "Unsupported accept header")
}

func (t GinLogin) Logout(c *gin.Context) {
	fmt.Println("Logout handler called")
	accept := c.Request.Header["Accept"][0]
	if accept == "text/html" {
		fmt.Println("Logging out client")
		refreshToken, err := c.Cookie("refreshToken")
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		_ = t.Keycloak.LogoutUser(refreshToken)
		c.SetCookie("accessToken", "", -1, "/", "", false, false)
		c.SetCookie("refreshToken", "", -1, "/", "", false, false)
		c.SetCookie("idToken", "", -1, "/", "", false, false)
		c.Header("HX-Redirect", "/")
		c.Status(http.StatusOK)
		return
	}
}
