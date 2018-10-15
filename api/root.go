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

// Service encapsulates API service operations
type Service struct {
	DB *data.Manager
}

// ErrorResponse represents an API response
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

//	Used to send back an error:
func sendErrorResponse(rw http.ResponseWriter, err error, code int) {
	//	Our return value
	response := ErrorResponse{
		Status:  code,
		Message: "Error: " + err.Error()}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(code)
	json.NewEncoder(rw).Encode(response)
}

// ShowUI redirects to the /ui/ url path
func ShowUI(rw http.ResponseWriter, req *http.Request) {
	// http.Redirect(rw, req, "/ui/", 301)
	fmt.Fprintf(rw, "Hello, world - UI")
}
