package controller

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"pdf_service_web/keycloak"
)

type Middleware struct {
	Keycloak keycloak.Keycloak
}

var AccessTokenKey = "accessToken"
var IdTokenKey = "idToken"
var RefreshTokenKey = "refreshToken"
var InvalidRefreshToken = errors.New("refresh token was not provided or invalid")

func (t Middleware) getAccessTokenUsingRefreshToken(c *gin.Context) (string, error) {
	refreshToken, err := c.Cookie(RefreshTokenKey)
	if err != nil {
		return "", err
	}

	token, err := t.Keycloak.SendLoginAuthAttemptWithRefreshToken(refreshToken)
	if err != nil {
		return "", InvalidRefreshToken
	}

	fmt.Println("Refreshed token")
	c.SetCookie("accessToken", token.AccessToken, token.AccessExpiresIn, "/", "", false, false)
	c.SetCookie("refreshToken", token.RefreshToken, token.RefreshExpiresIn, "/", "", false, false)
	c.SetCookie("idToken", token.IdToken, token.RefreshExpiresIn, "/", "", false, false)

	return token.AccessToken, nil
}

func (t Middleware) Authenticated(c *gin.Context) {
	onFailure := func() {
		c.SetCookie("accessToken", "", -1, "", "", false, false)
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	token, err := c.Request.Cookie(AccessTokenKey)
	if err != nil {
		accessToken, err := t.getAccessTokenUsingRefreshToken(c)
		if err != nil {
			fmt.Printf(err.Error())
			onFailure()
			c.Abort()
		}

		fmt.Println("Request authenticated")
		c.Set(AccessTokenKey, accessToken)
		c.Next()
		return
	}

	authenticated, err := t.authenticateJwtToken(token.Value)
	if err != nil {
		accessToken, err := t.getAccessTokenUsingRefreshToken(c)
		if err != nil {
			fmt.Printf(err.Error())
			onFailure()
			c.Abort()
		}

		fmt.Println("Request authenticated")
		c.Set(AccessTokenKey, accessToken)
		c.Next()
		return
	}

	if !authenticated {
		onFailure()
		c.Abort()
		return
	}

	fmt.Println("Request authenticated")
	c.Set(AccessTokenKey, token.Value)
	c.Next()
}

func (t Middleware) authenticateJwtToken(token string) (bool, error) {
	pem, err := jwt.ParseRSAPublicKeyFromPEM([]byte(t.Keycloak.GetSigningKey()))
	if err != nil {
		return false, err
	}

	tempClaim := &jwt.RegisteredClaims{}
	withClaims, err := jwt.ParseWithClaims(token, tempClaim, func(token *jwt.Token) (any, error) {
		return pem, nil
	})
	if err != nil {
		return false, err
	}

	if !withClaims.Valid {
		return false, nil
	}

	return true, nil
}
