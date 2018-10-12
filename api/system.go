package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/danesparza/iamserver/data"
)

// SystemResponse is a response for a system request
type SystemResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// NewUserRequest is a request to create a new user.  This is really the only type of requests
//	that requires a specific request type -- because a new user request includes a password
type NewUserRequest struct {
	Password string    `json:"password"`
	User     data.User `json:"user"`
}

// ShowUI redirects to the /ui/ url path
func ShowUI(rw http.ResponseWriter, req *http.Request) {
	// http.Redirect(rw, req, "/ui/", 301)
	fmt.Fprintf(rw, "Hello, world - UI")
}

// AddUser adds a user.  If the bearer token is not authorized for the operation, StatusUnauthorized is returned
func (service Service) AddUser(rw http.ResponseWriter, req *http.Request) {

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
	request := NewUserRequest{}
	err = json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusBadRequest)
		return
	}

	//	Perform the action with the context user
	newUser, err := service.DB.AddUser(user, request.User, request.Password)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusUnauthorized)
		return
	}

	//	Create our response and send information back:
	response := SystemResponse{
		Status:  http.StatusOK,
		Message: "User added",
		Data:    newUser,
	}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
}

// AddGroup adds a group.  If the bearer token is not authorized for the operation, StatusUnauthorized is returned
func (service Service) AddGroup(rw http.ResponseWriter, req *http.Request) {

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
	request := data.Group{}
	err = json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusBadRequest)
		return
	}

	//	Perform the action with the context user
	newGroup, err := service.DB.AddGroup(user, request.Name, request.Description)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusUnauthorized)
		return
	}

	//	Create our response and send information back:
	response := SystemResponse{
		Status:  http.StatusOK,
		Message: "Group added",
		Data:    newGroup,
	}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
}
