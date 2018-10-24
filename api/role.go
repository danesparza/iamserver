package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/danesparza/iamserver/data"
	"github.com/gorilla/mux"
)

// AddRole adds a role.  If the bearer token is not authorized for the operation, StatusUnauthorized is returned
func (service Service) AddRole(rw http.ResponseWriter, req *http.Request) {

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
	request := data.Role{}
	err = json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusBadRequest)
		return
	}

	//	Perform the action with the context user
	dataResponse, err := service.DB.AddRole(user, request.Name, request.Description)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	//	Create our response and send information back:
	response := SystemResponse{
		Status:  http.StatusOK,
		Message: "Role added",
		Data:    dataResponse,
	}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
}

// GetRole gets role information.  If the bearer token is not authorized for the operation, StatusUnauthorized is returned
func (service Service) GetRole(rw http.ResponseWriter, req *http.Request) {

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

	//	Parse the request
	vars := mux.Vars(req)

	//	Perform the action with the context user
	dataResponse, err := service.DB.GetRole(user, vars["rolename"])
	if err != nil {
		sendErrorResponse(rw, err, http.StatusNotFound)
		return
	}

	//	Create our response and send information back:
	response := SystemResponse{
		Status:  http.StatusOK,
		Message: "Role fetched",
		Data:    dataResponse,
	}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
}
