package api

import (
	"fmt"
	"net/http"
)

// OAuthRequest is an OAuth2 based request.  For more information on the
// various grant types that can use this request object:
// https://alexbilbie.com/guide-to-oauth-2-grants/
type OAuthRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scope        string `json:"scope"`
	UserName     string `json:"username"`
	Password     string `json:"password"`
	CSRFToken    string `json:"state"`
	RedirectURI  string `json:"redirect_uri"`
	ResponseType string `json:"response_type"`
	Code         string `json:"code"`
}

// OAuthResponse is an OAuth2 based response
type OAuthResponse struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    string `json:"expires_in"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// HelloWorld emits a hello world
func HelloWorld(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "Hello, world - service")
}
