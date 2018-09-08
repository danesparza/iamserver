package data

// Group represents a named collection of users
type Group struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Users       []string `json:"users"`
}
