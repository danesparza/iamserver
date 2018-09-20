package data

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/danesparza/badger"
	"github.com/xtgo/set"
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
	Policies    []string    `json:"policies"`
	Users       []string    `json:"users"`
	Groups      []string    `json:"groups"`
}

// AddRole adds a role to the system
func (store Manager) AddRole(context User, roleName string, roleDescription string) (Role, error) {
	//	Our return item
	retval := Role{}

	//	Our new role:
	role := Role{
		Name:        roleName,
		Description: roleDescription,
		Created:     time.Now(),
		CreatedBy:   context.Name,
		Updated:     time.Now(),
		UpdatedBy:   context.Name,
	}

	//	First -- does the role exist already?
	err := store.systemdb.View(func(txn *badger.Txn) error {
		_, err := txn.Get(GetKey("Role", role.Name))
		return err
	})

	//	If we didn't get an error, we have a problem:
	if err == nil {
		return retval, fmt.Errorf("Role already exists")
	}

	//	Serialize to JSON format
	encoded, err := json.Marshal(role)
	if err != nil {
		return retval, fmt.Errorf("Problem serializing the data: %s", err)
	}

	//	Save it to the database:
	err = store.systemdb.Update(func(txn *badger.Txn) error {
		err := txn.Set(GetKey("Role", role.Name), encoded)
		return err
	})

	//	If there was an error saving the data, report it:
	if err != nil {
		return retval, fmt.Errorf("Problem saving the data: %s", err)
	}

	//	Set our retval:
	retval = role

	//	Return our data:
	return retval, nil
}

// GetRole gets a user from the system
func (store Manager) GetRole(context User, roleName string) (Role, error) {
	//	Our return item
	retval := Role{}

	err := store.systemdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(GetKey("Role", roleName))
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}

		if len(val) > 0 {
			//	Unmarshal data into our item
			if err := json.Unmarshal(val, &retval); err != nil {
				return err
			}
		}

		return nil
	})

	//	If there was an error, report it:
	if err != nil {
		return retval, fmt.Errorf("Problem getting the data: %s", err)
	}

	//	Return our data:
	return retval, nil
}

// GetAllRoles gets all roles in the system
func (store Manager) GetAllRoles(context User) ([]Role, error) {
	//	Our return item
	retval := []Role{}

	err := store.systemdb.View(func(txn *badger.Txn) error {

		//	Get an iterator
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		//	Set our prefix
		prefix := GetKey("Role")

		//	Iterate over our values:
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {

			//	Get the item
			item := it.Item()

			//	Get the item key
			// k := item.Key()

			//	Get the item value
			val, err := item.Value()
			if err != nil {
				return err
			}

			if len(val) > 0 {
				//	Create our item:
				item := Role{}

				//	Unmarshal data into our item
				if err := json.Unmarshal(val, &item); err != nil {
					return err
				}

				//	Add to the array of returned users:
				retval = append(retval, item)
			}
		}
		return nil
	})

	//	If there was an error, report it:
	if err != nil {
		return retval, fmt.Errorf("Problem getting the list of items: %s", err)
	}

	//	Return our data:
	return retval, nil
}

// AddPoliciesToRole adds policies to a role -- and tracks that relationship
// at the role level and at the policy level
func (store Manager) AddPoliciesToRole(context User, roleName string, policies ...string) (Role, error) {
	//	Our return item
	retval := Role{}
	affectedPolicies := []Policy{}

	//	First -- validate that the role exists
	err := store.systemdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(GetKey("Role", roleName))
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}

		if len(val) > 0 {
			//	Unmarshal data into our item
			if err := json.Unmarshal(val, &retval); err != nil {
				return err
			}
		}

		return err
	})

	if err != nil {
		return retval, fmt.Errorf("Role does not exist")
	}

	//	Next:
	//	- Validate that each of the policies exist
	//	- Track each user in 'affectedPolicies'
	for _, currentpolicy := range policies {
		err := store.systemdb.View(func(txn *badger.Txn) error {
			item, err := txn.Get(GetKey("Policy", currentpolicy))

			if err != nil {
				return err
			}

			//	Deserialize the user and add to the list of affected users
			val, err := item.Value()
			if err != nil {
				return err
			}

			if len(val) > 0 {
				currentpolicyObject := Policy{}

				//	Unmarshal data into our item
				if err := json.Unmarshal(val, &currentpolicyObject); err != nil {
					return err
				}

				//	Add the object to our list of affected policies:
				affectedPolicies = append(affectedPolicies, currentpolicyObject)
			}

			return err
		})

		if err != nil {
			return retval, fmt.Errorf("Policy %s doesn't exist", currentpolicy)
		}
	}

	//	Get the roles's new list of policies from a merged (and deduped) list of:
	//	- The existing role policies
	// 	- The list of policies passed in
	allRolePolicies := append(retval.Policies, policies...)
	allUniqueRolePolicies := sort.StringSlice(allRolePolicies)

	sort.Sort(allUniqueRolePolicies)                  // sort the data first
	n := set.Uniq(allUniqueRolePolicies)              // Uniq returns the size of the set
	allUniqueRolePolicies = allUniqueRolePolicies[:n] // trim the duplicate elements

	//	Then add each of policies to both ...
	//	the role
	retval.Policies = allUniqueRolePolicies

	//	Serialize the role to JSON format
	encoded, err := json.Marshal(retval)
	if err != nil {
		return retval, fmt.Errorf("Problem serializing the data: %s", err)
	}

	//	Save it to the database:
	err = store.systemdb.Update(func(txn *badger.Txn) error {
		err := txn.Set(GetKey("Role", retval.Name), encoded)
		if err != nil {
			return err
		}

		//	Save each affected policy to the database (as part of the same transaction):
		for _, affectedPolicy := range affectedPolicies {

			//	and add the role to each policy
			currentRoles := append(affectedPolicy.Roles, roleName)
			allUniqueCurrentRoles := sort.StringSlice(currentRoles)

			sort.Sort(allUniqueCurrentRoles)                   // sort the data first
			cn := set.Uniq(allUniqueCurrentRoles)              // Uniq returns the size of the set
			allUniqueCurrentRoles = allUniqueCurrentRoles[:cn] // trim the duplicate elements
			affectedPolicy.Roles = allUniqueCurrentRoles

			encoded, err := json.Marshal(affectedPolicy)
			if err != nil {
				return err
			}

			err = txn.Set(GetKey("Policy", affectedPolicy.Name), encoded)
			if err != nil {
				return err
			}
		}

		return nil
	})

	//	If there was an error saving the data, report it:
	if err != nil {
		return retval, fmt.Errorf("Problem completing the role updates: %s", err)
	}

	//	Return our data:
	return retval, nil
}

// Attach role to user

// Attach role to group
