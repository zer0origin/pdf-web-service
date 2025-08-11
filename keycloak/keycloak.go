package keycloak

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
)

type UnauthenticatedUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

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

type AuthenticatedUser struct {
	Uid               string   `json:"sub"`
	EmailVerified     bool     `json:"email_verified"`
	Name              string   `json:"name"`
	PreferredUsername string   `json:"preferred_username"`
	GivenName         string   `json:"given_name"`
	FamilyName        string   `json:"family_name"`
	Email             string   `json:"email"`
	Organization      []string `json:"organization"`
}

type Keycloak struct {
	RealmHandler
	AdminHandler
}

var AccessTokenKey = "accessToken"
var IdTokenKey = "idToken"
var RefreshTokenKey = "refreshToken"
var InvalidToken = errors.New("token was not provided or invalid")

func (t Keycloak) AuthenticateJwtToken(token string) (jwt.Token, error) {
	pem, err := jwt.ParseRSAPublicKeyFromPEM([]byte(t.GetSigningKey()))
	if err != nil {
		return jwt.Token{}, err
	}

	tempClaim := &jwt.RegisteredClaims{}
	withClaims, err := jwt.ParseWithClaims(token, tempClaim, func(token *jwt.Token) (any, error) {
		return pem, nil
	})
	if err != nil {
		return jwt.Token{}, err
	}

	if !withClaims.Valid {
		return *withClaims, InvalidToken
	}

	return *withClaims, nil
}
