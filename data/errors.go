package data

import (
	"net/http"

	"github.com/pkg/errors"
)

var (
	// ErrRequestDenied is returned when an access request can not be satisfied by any policy.
	ErrRequestDenied = &errorWithContext{
		error:  errors.New("Request was denied by default"),
		code:   http.StatusForbidden,
		status: http.StatusText(http.StatusForbidden),
		reason: "The request was denied because no matching policy was found.",
	}

	// ErrRequestForcefullyDenied is returned when an access request is explicitly denied by a policy.
	ErrRequestForcefullyDenied = &errorWithContext{
		error:  errors.New("Request was forcefully denied"),
		code:   http.StatusForbidden,
		status: http.StatusText(http.StatusForbidden),
		reason: "The request was denied because a policy denied request.",
	}

	// ErrNotFound is returned when a resource can not be found.
	ErrNotFound = &errorWithContext{
		error:  errors.New("Resource could not be found"),
		code:   http.StatusNotFound,
		status: http.StatusText(http.StatusNotFound),
	}
)

type errorWithContext struct {
	code   int
	reason string
	status string
	error
}

// StatusCode returns the status code of this error.
func (e *errorWithContext) StatusCode() int {
	return e.code
}

// RequestID returns the ID of the request that caused the error, if applicable.
func (e *errorWithContext) RequestID() string {
	return ""
}

// Reason returns the reason for the error, if applicable.
func (e *errorWithContext) Reason() string {
	return e.reason
}

// ID returns the error id, if applicable.
func (e *errorWithContext) Status() string {
	return e.status
}

// Details returns details on the error, if applicable.
func (e *errorWithContext) Details() []map[string]interface{} {
	return []map[string]interface{}{}
}
