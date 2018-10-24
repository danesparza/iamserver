package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/danesparza/iamserver/data"
	"github.com/gorilla/mux"
)

// AddPolicy adds a policy.  If the bearer token is not authorized for the operation, StatusUnauthorized is returned
func (service Service) AddPolicy(rw http.ResponseWriter, req *http.Request) {

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
	request := data.Policy{}
	err = json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusBadRequest)
		return
	}

	//	Perform the action with the context user
	dataResponse, err := service.DB.AddPolicy(user, request)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	//	Create our response and send information back:
	response := SystemResponse{
		Status:  http.StatusCreated,
		Message: "Policy added",
		Data:    dataResponse,
	}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
}

// GetPolicy gets a policy.  If the bearer token is not authorized for the operation, StatusUnauthorized is returned
func (service Service) GetPolicy(rw http.ResponseWriter, req *http.Request) {

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
	dataResponse, err := service.DB.GetPolicy(user, vars["policyname"])
	if err != nil {
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	//	Create our response and send information back:
	response := SystemResponse{
		Status:  http.StatusOK,
		Message: "Policy fetched",
		Data:    dataResponse,
	}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
}

// GetAllPolicies gets all policies in the system.  If the bearer token is not authorized for the operation, StatusUnauthorized is returned
func (service Service) GetAllPolicies(rw http.ResponseWriter, req *http.Request) {

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

	//	Perform the action with the context user
	dataResponse, err := service.DB.GetAllPolicies(user)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	//	Create our response and send information back:
	response := SystemResponse{
		Status:  http.StatusOK,
		Message: "Policies fetched",
		Data:    dataResponse,
	}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
}

// AttachPolicyToUsers attaches user(s) to a policy.  If the bearer token is not authorized for the operation, StatusUnauthorized is returned
func (service Service) AttachPolicyToUsers(rw http.ResponseWriter, req *http.Request) {

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

	//	Get our list of users:
	users := vars["userlist"]
	userList := strings.Split(users, ",")

	//	Perform the action with the context user
	dataResponse, err := service.DB.AttachPolicyToUsers(user, vars["policyname"], userList...)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	//	Create our response and send information back:
	response := SystemResponse{
		Status:  http.StatusOK,
		Message: "Attached policy to user(s)",
		Data:    dataResponse,
	}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
}

// AttachPolicyToGroups attaches group(s) to a policy.  If the bearer token is not authorized for the operation, StatusUnauthorized is returned
func (service Service) AttachPolicyToGroups(rw http.ResponseWriter, req *http.Request) {

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

	//	Get our list of groups:
	groups := vars["grouplist"]
	groupList := strings.Split(groups, ",")

	//	Perform the action with the context user
	dataResponse, err := service.DB.AttachPolicyToGroups(user, vars["policyname"], groupList...)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	//	Create our response and send information back:
	response := SystemResponse{
		Status:  http.StatusOK,
		Message: "Attached policy to group(s)",
		Data:    dataResponse,
	}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
}

// GetPoliciesForUser gets all policies for a given user.  If the bearer token is not authorized for the operation, StatusUnauthorized is returned
func (service Service) GetPoliciesForUser(rw http.ResponseWriter, req *http.Request) {

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
	dataResponse, err := service.DB.GetPoliciesForUser(user, vars["username"])
	if err != nil {
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	//	Create our response and send information back:
	response := SystemResponse{
		Status:  http.StatusOK,
		Message: "Policies in effect for user",
		Data:    dataResponse,
	}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
}
