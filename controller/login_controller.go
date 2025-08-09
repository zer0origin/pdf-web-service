package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pdf_service_web/keycloak"
)

type LoginController struct {
	AuthenticatedRedirect string
	RealmConfig           keycloak.RealmConfig
}

func (t LoginController) LoginRender(c *gin.Context) {
	if refreshTokenCookie, err := c.Request.Cookie("refreshToken"); err == nil {
		if _, err := c.Request.Cookie("accessToken"); err == nil {
			c.Redirect(http.StatusFound, t.AuthenticatedRedirect)
			return
		}

		token, err := t.RealmConfig.SendLoginAuthAttemptWithRefreshToken(refreshTokenCookie.Value)
		if err == nil {
			c.SetCookie("accessToken", token.AccessToken, token.AccessExpiresIn, "/", "", false, false)
			c.SetCookie("refreshToken", token.RefreshToken, token.RefreshExpiresIn, "/", "", false, false)
			c.SetCookie("idToken", token.IdToken, token.RefreshExpiresIn, "/", "", false, false)
			c.Redirect(http.StatusFound, t.AuthenticatedRedirect)
			return
		}
	}

	c.HTML(http.StatusOK, "login", gin.H{})
}

func (t LoginController) LoginAuthHandler(c *gin.Context) {
	accept := c.Request.Header["Accept"][0]
	if accept == "text/html" {
		username, userPresent := c.GetPostForm("username")
		password, passPresent := c.GetPostForm("password")

		if !userPresent || !passPresent {
			c.JSON(http.StatusUnprocessableEntity, gin.H{})
			return
		}

		authUser, err := t.RealmConfig.SendLoginAuthAttemptWithPasswordAndUsername(username, password)
		if err != nil {
			c.JSON(http.StatusBadRequest, err) //TODO RETURN ERROR AS HTML
			return
		}

		c.SetCookie("accessToken", authUser.AccessToken, authUser.AccessExpiresIn, "/", "", false, false)
		c.SetCookie("refreshToken", authUser.RefreshToken, authUser.RefreshExpiresIn, "/", "", false, false)
		c.SetCookie("idToken", authUser.IdToken, authUser.RefreshExpiresIn, "/", "", false, false)
		c.Header("HX-Redirect", t.AuthenticatedRedirect)
		c.Status(http.StatusOK)
		return
	}

	if accept == "application/json" || accept == "*/*" {
		loginInfo := &keycloak.UnauthenticatedUser{}
		err := c.ShouldBindJSON(loginInfo)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		authAttempt, err := t.RealmConfig.SendLoginAuthAttemptWithPasswordAndUsername(loginInfo.Username, loginInfo.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		c.JSON(http.StatusOK, authAttempt)
		return
	}

	c.JSON(http.StatusUnprocessableEntity, gin.H{})
}
