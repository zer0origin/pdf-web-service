package keycloak

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	AccessExpiresIn  int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	IdToken          string `json:"id_token"`
	NotBeforePolicy  int    `json:"not_before_policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

type Api struct {
	*RealmHandler
	AdminHandler
}

var AccessTokenKey = "accessToken"
var IdTokenKey = "idToken"
var RefreshTokenKey = "refreshToken"
var InvalidToken = errors.New("token was not provided or invalid")

// AuthenticateJwtToken Parse a token after validating its authenticity, if invalid then return a blank token.
func (t Api) AuthenticateJwtToken(token string) (jwt.Token, error) {
	keyStr, err := t.GetSigningKey()
	if err != nil {
		return jwt.Token{}, fmt.Errorf("failed to get public key from keyclaok server: %s %s", err.Error(), token)
	}

	pem, err := jwt.ParseRSAPublicKeyFromPEM([]byte(keyStr))
	if err != nil {
		return jwt.Token{}, fmt.Errorf("failed to parse token: %s %s", err.Error(), token)
	}

	tempClaim := &jwt.RegisteredClaims{}
	withClaims, err := jwt.ParseWithClaims(token, tempClaim, func(token *jwt.Token) (any, error) {
		return pem, nil
	})
	if err != nil {
		return jwt.Token{}, fmt.Errorf("failed to parse token: %s %s", err.Error(), token)
	}

	if !withClaims.Valid {
		return *withClaims, InvalidToken
	}
	return *withClaims, nil
}

// ParseTokenUnverified Assume the token is authentic and parse it without checking authenticity.
func (t Api) ParseTokenUnverified(token string) (jwt.Token, error) {
	tempClaim := &jwt.RegisteredClaims{}
	unverified, _, err := jwt.NewParser().ParseUnverified(token, tempClaim)
	if err != nil {
		return jwt.Token{}, err
	}

	return *unverified, nil
}
