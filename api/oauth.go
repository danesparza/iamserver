package api

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

// AuthRequest is an OAuth2 based request.  For more information on the
// various grant types that can use this request object:
// https://alexbilbie.com/guide-to-oauth-2-grants/
type AuthRequest struct {
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

// AuthResponse is an OAuth2 based response
type AuthResponse struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    string `json:"expires_in"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// HelloWorld emits a hello world
func HelloWorld(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "Hello, world - service")
}

// authHeaderValid returns true if the passed header value is a valid
// for a "bearer token" authorization field -- otherwise return false
func authHeaderValid(header string) bool {
	retval := true

	//	If we don't have at least x number characters,
	//	it must not include the prefix text 'Bearer '
	if len(header) < len("Bearer ") {
		return false
	}

	//	If the first part of the string isn't 'Bearer ' then it's not a bearer token...
	if strings.EqualFold(header[:len("Bearer ")], "Bearer ") != true {
		return false
	}

	return retval
}

// basicHeaderValid returns true if the passed header value is a valid
// for a "http basic authentication" authorization field -- otherwise return false
func basicHeaderValid(header string) bool {
	retval := true

	//	If we don't have at least x number characters,
	//	it must not include the prefix text 'Basic '
	if len(header) < len("Basic ") {
		return false
	}

	//	If the first part of the string isn't 'Basic ' then it's not a basic auth header...
	if strings.EqualFold(header[:len("Basic ")], "Basic ") != true {
		return false
	}

	return retval
}

// getTokenFromAuthHeader returns the token itself from the Authorization header
func getTokenFromAuthHeader(header string) string {
	retval := ""

	//	If we don't have at least x number characters,
	//	it must not include the prefix text 'Bearer '
	if len(header) < len("Bearer ") {
		return ""
	}

	//	If the first part of the string isn't 'Bearer ' then it's not a bearer token...
	if strings.EqualFold(header[:len("Bearer ")], "Bearer ") != true {
		return ""
	}

	//	Get the token and decode it
	encodedToken := header[len("Bearer "):]
	tokenBytes, err := base64.StdEncoding.DecodeString(encodedToken)
	if err != nil {
		return ""
	}

	//	Change the type to string
	retval = string(tokenBytes)

	return retval
}

// getCredentialsFromAuthHeader returns the username/password from the Authorization header
func getCredentialsFromAuthHeader(header string) (string, string) {
	username := ""
	password := ""

	//	If we don't have at least x number characters,
	//	it must not include the prefix text 'Basic '
	if len(header) < len("Basic ") {
		return "", ""
	}

	//	If the first part of the string isn't 'Basic ' then it's not a basic auth string...
	if strings.EqualFold(header[:len("Basic ")], "Basic ") != true {
		return "", ""
	}

	//	Get the credentials and decode them
	encodedCredentials := header[len("Basic "):]
	credentialBytes, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		return "", ""
	}

	//	Change the type to string
	credentials := strings.Split(string(credentialBytes), ":")

	username = credentials[0]
	password = credentials[1]

	return username, password
}
