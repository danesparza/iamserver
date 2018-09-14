package data

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/danesparza/iamserver/policy"

	"github.com/danesparza/badger"
	null "gopkg.in/guregu/null.v3"
	"gopkg.in/guregu/null.v3/zero"
)

// Policy is an AWS style policy document.  They wrap up the following ideas:
// - Resources: The things in a system that users would need permissions to
// - Actions: The interactions users have with those resources
// - Effects: The permissive effects of a policy (allow or deny)
// - Conditions: Additional information to take into account when evaluating a policy
// Policies can be attached to a user or a user group
type Policy struct {
	Name      string      `json:"sid"`
	Effect    string      `json:"effect"`
	Resources []string    `json:"resources"`
	Actions   []string    `json:"actions"`
	Roles     []string    `json:"roles"`
	Users     []string    `json:"users"`
	Groups    []string    `json:"groups"`
	Created   time.Time   `json:"created"`
	CreatedBy string      `json:"created_by"`
	Updated   time.Time   `json:"updated"`
	UpdatedBy string      `json:"updated_by"`
	Deleted   zero.Time   `json:"deleted"`
	DeletedBy null.String `json:"deleted_by"`
}

// AddPolicy adds a policy to the system
func (store Manager) AddPolicy(context User, newPolicy Policy) (Policy, error) {
	//	Our return item
	retval := Policy{}

	//	First -- does the policy exist already?
	err := store.systemdb.View(func(txn *badger.Txn) error {
		_, err := txn.Get(GetKey("Policy", newPolicy.Name))
		return err
	})

	//	If we didn't get an error, we have a problem:
	if err == nil {
		return retval, fmt.Errorf("Policy already exists")
	}

	//	Check Effect
	if (newPolicy.Effect != policy.Allow) && (newPolicy.Effect != policy.Deny) {
		return retval, fmt.Errorf("Policy must have 'allow' or 'deny' effect")
	}

	// 	Check Resources / Actions (they can't be blank or empty)
	if len(newPolicy.Resources) == 0 || len(newPolicy.Actions) == 0 {
		return retval, fmt.Errorf("Policy must have 'resources' and 'actions' associated with it")
	}

	//	Associated resources have to exist

	//	Make sure when adding a new policy, users / roles / groups are empty:
	newPolicy.Users = []string{}
	newPolicy.Roles = []string{}
	newPolicy.Groups = []string{}

	//	Update the created / updated fields:
	newPolicy.Created = time.Now()
	newPolicy.Updated = time.Now()
	newPolicy.CreatedBy = context.Name
	newPolicy.UpdatedBy = context.Name

	//	Serialize to JSON format
	encoded, err := json.Marshal(newPolicy)
	if err != nil {
		return retval, fmt.Errorf("Problem serializing the data: %s", err)
	}

	//	Save it to the database:
	err = store.systemdb.Update(func(txn *badger.Txn) error {
		err := txn.Set(GetKey("Policy", newPolicy.Name), encoded)
		return err
	})

	//	If there was an error saving the data, report it:
	if err != nil {
		return retval, fmt.Errorf("Problem saving the data: %s", err)
	}

	//	Set our retval:
	retval = newPolicy

	//	Return our data:
	return retval, nil
}

// Add policy to role

// Attach policy to user

// Attach policy to group
