package controller

import (
	"fmt"
	"net/http"
	"pdf_service_web/keycloak"

	"github.com/gin-gonic/gin"
)

type GinMiddleware struct {
	Keycloak *keycloak.Api
}

func (t GinMiddleware) getAccessTokenUsingRefreshToken(c *gin.Context) (string, error) {
	refreshToken, err := c.Cookie(keycloak.RefreshTokenKey)
	if err != nil {
		return "", err
	}

	token, err := t.Keycloak.SendLoginAuthAttemptWithRefreshToken(refreshToken)
	if err != nil {
		return "", keycloak.InvalidToken
	}

	fmt.Println("Refreshed token")
	c.SetCookie("accessToken", token.AccessToken, token.AccessExpiresIn, "/", "", false, false)
	c.SetCookie("refreshToken", token.RefreshToken, token.RefreshExpiresIn, "/", "", false, false)
	c.SetCookie("idToken", token.IdToken, token.RefreshExpiresIn, "/", "", false, false)

	return token.AccessToken, nil
}

func (t GinMiddleware) RequireAuthenticated(c *gin.Context) {
	onFailure := func() {
		c.SetCookie("accessToken", "", -1, "", "", false, false)
		c.Redirect(http.StatusTemporaryRedirect, "/") //Login page
		c.Abort()
	}

	onSucceeded := func(accessToken string) {
		c.Set(keycloak.AccessTokenKey, accessToken)
		c.Next()
	}

	err := t.IsAuthenticated(c, true, onFailure, onSucceeded)
	if err != nil {
		onFailure()
		return
	}
}

func (t GinMiddleware) IsAuthenticated(c *gin.Context, destroyCookies bool, onFailure func(), onSucceeded func(accessToken string)) error {
	destroyCookiesFunc := func() {
		if destroyCookies {
			c.SetCookie(keycloak.AccessTokenKey, "", -1, "", "", false, false)
			c.SetCookie(keycloak.RefreshTokenKey, "", -1, "", "", false, false)
		}
	}

	token, err := c.Request.Cookie(keycloak.AccessTokenKey)
	if err != nil {
		accessToken, err := t.getAccessTokenUsingRefreshToken(c)
		if err != nil {
			onFailure()
			destroyCookiesFunc()
			return err
		}

		onSucceeded(accessToken)
		return nil
	}

	_, err = t.Keycloak.AuthenticateJwtToken(token.Value)
	if err != nil {
		accessToken, err := t.getAccessTokenUsingRefreshToken(c)
		if err != nil {
			onFailure()
			destroyCookiesFunc()
			return err
		}

		onSucceeded(accessToken)
		return nil
	}

	onSucceeded(token.Value)
	return nil
}
