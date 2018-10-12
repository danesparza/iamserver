package data

// Request represents a request to be validated
type Request struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}
