package data

// ResourceAction represents an action that can be
// performed in relation to the parent resource.
// Example: list, get, update
type ResourceAction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
