package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/danesparza/iamserver/data"
)

// TokenResponse is a response for a bearer token
type TokenResponse struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   string `json:"expires_in"`
	AccessToken string `json:"access_token"`
}

// AuthResponse is a response structure returned after validating a request
type AuthResponse struct {
	Authorized bool `json:"authorized"`
}

// GetTokenForCredentials gets a bearer token for a given set of credentials
func (service Service) GetTokenForCredentials(rw http.ResponseWriter, req *http.Request) {

	//	req.Body is a ReadCloser -- we need to remember to close it:
	defer req.Body.Close()

	//	Get the authorization header:
	authHeader := req.Header.Get("Authorization")

	//	If the basic auth header wasn't supplied, return an error
	if basicHeaderValid(authHeader) != true {
		sendErrorResponse(rw, fmt.Errorf("HTTP basic auth credentials not supplied"), http.StatusUnauthorized)
		return
	}

	//	Get just the credentials from basic auth information:
	clientid, clientsecret := getCredentialsFromAuthHeader(authHeader)

	//	Get the user from the credentials:
	user, err := service.DB.GetUserWithCredentials(clientid, clientsecret)
	if err != nil {
		sendErrorResponse(rw, fmt.Errorf("HTTP basic auth credentials can't be retrieved"), http.StatusUnauthorized)
		return
	}

	//	If the user has enabled two factor authentication, make sure they authenticate using two-factor:

	//	Get a token for a user:
	tokenttlstring := viper.GetString("apiservice.tokenttl")
	tokenttl, err := strconv.Atoi(tokenttlstring)
	if err != nil {
		sendErrorResponse(rw, fmt.Errorf("The apiservice.tokenttl configuration is invalid"), http.StatusUnprocessableEntity)
		return
	}
	token, err := service.DB.GetNewToken(user, time.Duration(tokenttl)*time.Minute)

	//	Create our response and send information back:
	encodedToken := base64.StdEncoding.EncodeToString([]byte(token.ID))
	response := TokenResponse{
		TokenType:   "Bearer",
		ExpiresIn:   strconv.FormatFloat(token.Expires.Sub(time.Now()).Seconds(), 'f', 0, 64),
		AccessToken: encodedToken,
	}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
}

// IsRequestAuthorized returns whether a request is authorized for a given bearer token and request object
func (service Service) IsRequestAuthorized(rw http.ResponseWriter, req *http.Request) {

	//	req.Body is a ReadCloser -- we need to remember to close it:
	defer req.Body.Close()

	//	Get the authorization header:
	authHeader := req.Header.Get("Authorization")

	//	If the auth header wasn't supplied, return an error
	if authHeaderValid(authHeader) != true {
		sendErrorResponse(rw, fmt.Errorf("Bearer token was not supplied"), http.StatusForbidden)
		return
	}

	//	Get just the bearer token itself:
	token := getBearerTokenFromAuthHeader(authHeader)

	//	Get the user from the token:
	user, err := service.DB.GetUserForToken(token)
	if err != nil {
		sendErrorResponse(rw, fmt.Errorf("Token not authorized or not valid"), http.StatusUnauthorized)
		return
	}

	//	Parse the request JSON
	request := data.Request{}
	err = json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusBadRequest)
		return
	}

	//	See if the request is valid
	authorized := service.DB.IsUserRequestAuthorized(user, &request)

	//	Create our response and send information back:
	response := AuthResponse{
		Authorized: authorized,
	}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
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

// getBearerTokenFromAuthHeader returns the token itself from the Authorization header
func getBearerTokenFromAuthHeader(header string) string {
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
