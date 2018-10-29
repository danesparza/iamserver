package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// TotpEnrollmentFinishRequest represents a request to
// complete the TOTP enrollment request.  A passcode (from the OTP device / authenticator app)
// is required to validate and complete the process
type TotpEnrollmentFinishRequest struct {
	PassCode string `json:"passcode"`
}

// BeginTOTPEnrollment begins TOTP (two factor auth) enrollment.
// If the bearer token is not authorized for the operation, StatusUnauthorized is returned
func (service Service) BeginTOTPEnrollment(rw http.ResponseWriter, req *http.Request) {

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
	_, err = service.DB.BeginTOTPEnrollment(user.Name, 1*time.Hour)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	//	Create our response and send information back:
	response := SystemResponse{
		Status:  http.StatusAccepted,
		Message: "Enrollment started",
		Data:    "",
	}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
}

// GetTOTPEnrollmentImage gets the TOTP (two factor auth) enrollment image.
// If the bearer token is not authorized for the operation, StatusUnauthorized is returned
func (service Service) GetTOTPEnrollmentImage(rw http.ResponseWriter, req *http.Request) {

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
	enrollment, err := service.DB.GetTOTPEnrollment(user.Name)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	//	Get the image:
	img, err := enrollment.GetImage()
	if err != nil {
		sendErrorResponse(rw, err, http.StatusNotFound)
		return
	}

	//	Return the image response:
	rw.Header().Set("Content-Type", "image/png")
	rw.Header().Set("Content-Length", strconv.Itoa(len(img)))
	rw.Write(img)
}

// FinishTOTPEnrollment finishes TOTP (two factor auth) enrollment by verifying the first code.
// If the bearer token is not authorized for the operation, StatusUnauthorized is returned
func (service Service) FinishTOTPEnrollment(rw http.ResponseWriter, req *http.Request) {

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
	request := TotpEnrollmentFinishRequest{}
	err = json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusBadRequest)
		return
	}

	//	Perform the action with the context user
	_, err = service.DB.FinishTOTPEnrollment(user.Name, request.PassCode)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	//	Create our response and send information back:
	response := SystemResponse{
		Status:  http.StatusAccepted,
		Message: "Enrollment completed",
		Data:    "",
	}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
}
