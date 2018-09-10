package data

import (
	"time"

	null "gopkg.in/guregu/null.v3"
	"gopkg.in/guregu/null.v3/zero"
)

// Role represents a named collection of policies.
// Roles can be attached to a user or a user group
type Role struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Created     time.Time   `json:"created"`
	CreatedBy   string      `json:"created_by"`
	Updated     time.Time   `json:"updated"`
	UpdatedBy   string      `json:"updated_by"`
	Deleted     zero.Time   `json:"deleted"`
	DeletedBy   null.String `json:"deleted_by"`
	Policies    []Policy    `json:"policies"`
	Users       []string    `json:"users"`
	Groups      []string    `json:"groups"`
}
