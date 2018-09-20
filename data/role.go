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

// AttachRoleToUsers attaches a role to the given user(s)
func (store Manager) AttachRoleToUsers(context User, roleName string, users ...string) (Role, error) {
	//	Our return item
	retval := Role{}
	affectedUsers := []User{}

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
	//	- Validate that each of the users exist
	//	- Track each user in 'affectedUsers'
	for _, currentuser := range users {
		err := store.systemdb.View(func(txn *badger.Txn) error {
			item, err := txn.Get(GetKey("User", currentuser))

			if err != nil {
				return err
			}

			//	Deserialize the user and add to the list of affected users
			val, err := item.Value()
			if err != nil {
				return err
			}

			if len(val) > 0 {
				currentuserObject := User{}

				//	Unmarshal data into our item
				if err := json.Unmarshal(val, &currentuserObject); err != nil {
					return err
				}

				//	Add the object to our list of affected users:
				affectedUsers = append(affectedUsers, currentuserObject)
			}

			return err
		})

		if err != nil {
			return retval, fmt.Errorf("User %s doesn't exist", currentuser)
		}
	}

	//	Get the role's new list of users from a merged (and deduped) list of:
	//	- The existing role users
	// 	- The list of users passed in
	allRoleUsers := append(retval.Users, users...)
	allUniqueRoleUsers := sort.StringSlice(allRoleUsers)

	sort.Sort(allUniqueRoleUsers)               // sort the data first
	n := set.Uniq(allUniqueRoleUsers)           // Uniq returns the size of the set
	allUniqueRoleUsers = allUniqueRoleUsers[:n] // trim the duplicate elements

	//	Then add each of users to the role
	retval.Users = allUniqueRoleUsers

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

		//	Save each affected user to the database (as part of the same transaction):
		for _, affectedUser := range affectedUsers {

			//	and add the group to each user
			currentRoles := append(affectedUser.Roles, roleName)
			allUniqueCurrentRoles := sort.StringSlice(currentRoles)

			sort.Sort(allUniqueCurrentRoles)                   // sort the data first
			cn := set.Uniq(allUniqueCurrentRoles)              // Uniq returns the size of the set
			allUniqueCurrentRoles = allUniqueCurrentRoles[:cn] // trim the duplicate elements
			affectedUser.Roles = allUniqueCurrentRoles

			encoded, err := json.Marshal(affectedUser)
			if err != nil {
				return err
			}

			err = txn.Set(GetKey("User", affectedUser.Name), encoded)
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

// AttachRoleToGroups attaches a role to the given group(s)
func (store Manager) AttachRoleToGroups(context User, roleName string, groups ...string) (Role, error) {
	//	Our return item
	retval := Role{}
	affectedGroups := []Group{}

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
	//	- Validate that each of the groups exist
	//	- Track each group in 'affectedGroups'
	for _, currentgroup := range groups {
		err := store.systemdb.View(func(txn *badger.Txn) error {
			item, err := txn.Get(GetKey("Group", currentgroup))

			if err != nil {
				return err
			}

			//	Deserialize the group and add to the list of affected groups
			val, err := item.Value()
			if err != nil {
				return err
			}

			if len(val) > 0 {
				currentgroupObject := Group{}

				//	Unmarshal data into our item
				if err := json.Unmarshal(val, &currentgroupObject); err != nil {
					return err
				}

				//	Add the object to our list of affected groups:
				affectedGroups = append(affectedGroups, currentgroupObject)
			}

			return err
		})

		if err != nil {
			return retval, fmt.Errorf("Group %s doesn't exist", currentgroup)
		}
	}

	//	Get the role's new list of groups from a merged (and deduped) list of:
	//	- The existing role groups
	// 	- The list of groups passed in
	allRoleGroups := append(retval.Groups, groups...)
	allUniqueRoleGroups := sort.StringSlice(allRoleGroups)

	sort.Sort(allUniqueRoleGroups)                // sort the data first
	n := set.Uniq(allUniqueRoleGroups)            // Uniq returns the size of the set
	allUniqueRoleGroups = allUniqueRoleGroups[:n] // trim the duplicate elements

	//	Then add each of the groups to the role
	retval.Groups = allUniqueRoleGroups

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

		//	Save each affected group to the database (as part of the same transaction):
		for _, affectedGroup := range affectedGroups {

			//	and add the group to each role
			currentRoles := append(affectedGroup.Roles, roleName)
			allUniqueCurrentRoles := sort.StringSlice(currentRoles)

			sort.Sort(allUniqueCurrentRoles)                   // sort the data first
			cn := set.Uniq(allUniqueCurrentRoles)              // Uniq returns the size of the set
			allUniqueCurrentRoles = allUniqueCurrentRoles[:cn] // trim the duplicate elements
			affectedGroup.Roles = allUniqueCurrentRoles

			encoded, err := json.Marshal(affectedGroup)
			if err != nil {
				return err
			}

			err = txn.Set(GetKey("Group", affectedGroup.Name), encoded)
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
