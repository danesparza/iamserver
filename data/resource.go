package data

import (
	"time"

	null "gopkg.in/guregu/null.v3"
	"gopkg.in/guregu/null.v3/zero"
)

// Resource represents a thing that can be acted on
type Resource struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Created     time.Time        `json:"created"`
	CreatedBy   string           `json:"created_by"`
	Updated     time.Time        `json:"updated"`
	UpdatedBy   string           `json:"updated_by"`
	Deleted     zero.Time        `json:"deleted"`
	DeletedBy   null.String      `json:"deleted_by"`
	Actions     []ResourceAction `json:"actions"`
}
