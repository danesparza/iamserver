package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hako/durafmt"

	"github.com/danesparza/iamserver/data"
)

// OverviewResponse represents a response to the GetOverview call
type OverviewResponse struct {
	SystemOverview  data.SystemOverview `json:"overview"`
	Uptime          string              `json:"uptime"`
	UserName        string              `json:"user_name"`
	UserDescription string              `json:"user_description"`
}

// GetOverview gets the system overview information.  If the bearer token is not authorized for the operation, StatusUnauthorized is returned
func (service Service) GetOverview(rw http.ResponseWriter, req *http.Request) {

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
	dataResponse, err := service.DB.GetOverview(user)
	if err != nil {
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	formattedUptime := durafmt.ParseShort(time.Since(service.StartTime)).String()

	//	Create our response and send information back:
	response := SystemResponse{
		Status:  http.StatusOK,
		Message: "Overview fetched",
		Data: OverviewResponse{
			SystemOverview:  dataResponse,
			Uptime:          formattedUptime,
			UserName:        user.Name,
			UserDescription: user.Description,
		},
	}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
}
