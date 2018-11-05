package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SearchRequest represents a request to search for information
type SearchRequest struct {
	SearchExpression string `json:"expression"`
}

// Search searches the system for a given user/group/role/policy/resource name using the supplied regex expression.  If the bearer token is not authorized for the operation, StatusUnauthorized is returned
func (service Service) Search(rw http.ResponseWriter, req *http.Request) {

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
	request := SearchRequest{}
	err = json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusBadRequest)
		return
	}

	//	Perform the action with the context user
	dataResponse, err := service.DB.Search(user, request.SearchExpression)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	//	Create our response and send information back:
	response := SystemResponse{
		Status:  http.StatusOK,
		Message: "Search results",
		Data:    dataResponse,
	}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
}
