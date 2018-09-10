package data

import (
	"time"

	null "gopkg.in/guregu/null.v3"
	"gopkg.in/guregu/null.v3/zero"
)

// PolicyEffect is the permissive effects of a policy (allow or deny)
type PolicyEffect struct {
	Allow string
	Deny  string
}

// Policy is an AWS style policy document.  They wrap up the following ideas:
// - Resources: The things in a system that users would need permissions to
// - Actions: The interactions users have with those resources
// - Effects: The permissive effects of a policy (allow or deny)
// - Conditions: Additional information to take into account when evaluating a policy
// Policies can be attached to a user or a user group
type Policy struct {
	SID       string           `json:"sid"`
	Effect    PolicyEffect     `json:"effect"`
	Resource  string           `json:"resource"`
	Actions   []ResourceAction `json:"actions"`
	Created   time.Time        `json:"created"`
	CreatedBy string           `json:"created_by"`
	Updated   time.Time        `json:"updated"`
	UpdatedBy string           `json:"updated_by"`
	Deleted   zero.Time        `json:"deleted"`
	DeletedBy null.String      `json:"deleted_by"`
	Roles     []string         `json:"roles"`
	Users     []string         `json:"users"`
	Groups    []string         `json:"groups"`
}
