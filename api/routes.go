package api

import (
	"encoding/json"
	"net/http"

	"github.com/danesparza/iamserver/data"
)

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
