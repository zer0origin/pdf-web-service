package keycloak

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
