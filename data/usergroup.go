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

// Group represents a named collection of users
type Group struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Created     time.Time   `json:"created"`
	CreatedBy   string      `json:"created_by"`
	Updated     time.Time   `json:"updated"`
	UpdatedBy   string      `json:"updated_by"`
	Deleted     zero.Time   `json:"deleted"`
	DeletedBy   null.String `json:"deleted_by"`
	Users       []string    `json:"users"`
	Policies    []string    `json:"policies"`
	Roles       []string    `json:"roles"`
}

// AddGroup adds a user group to the system
func (store Manager) AddGroup(context User, groupName string, groupDescription string) (Group, error) {
	//	Our return item
	retval := Group{}

	//	Our new group:
	group := Group{
		Name:        groupName,
		Description: groupDescription,
		Created:     time.Now(),
		CreatedBy:   context.Name,
		Updated:     time.Now(),
		UpdatedBy:   context.Name,
	}

	//	First -- does the group exist already?
	err := store.systemdb.View(func(txn *badger.Txn) error {
		_, err := txn.Get(GetKey("Group", group.Name))
		return err
	})

	//	If we didn't get an error, we have a problem:
	if err == nil {
		return retval, fmt.Errorf("Group already exists")
	}

	//	Serialize to JSON format
	encoded, err := json.Marshal(group)
	if err != nil {
		return retval, fmt.Errorf("Problem serializing the data: %s", err)
	}

	//	Save it to the database:
	err = store.systemdb.Update(func(txn *badger.Txn) error {
		err := txn.Set(GetKey("Group", group.Name), encoded)
		return err
	})

	//	If there was an error saving the data, report it:
	if err != nil {
		return retval, fmt.Errorf("Problem saving the data: %s", err)
	}

	//	Set our retval:
	retval = group

	//	Return our data:
	return retval, nil
}

// GetGroup gets a user from the system
func (store Manager) GetGroup(context User, groupName string) (Group, error) {
	//	Our return item
	retval := Group{}

	err := store.systemdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(GetKey("Group", groupName))
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

// GetAllGroups gets all groups in the system
func (store Manager) GetAllGroups(context User) ([]Group, error) {
	//	Our return item
	retval := []Group{}

	err := store.systemdb.View(func(txn *badger.Txn) error {

		//	Get an iterator
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		//	Set our prefix
		prefix := GetKey("Group")

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
				item := Group{}

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

// AddUsersToGroup adds user(s) to a group -- and tracks that relationship
// at the group level and at the user level
func (store Manager) AddUsersToGroup(context User, groupName string, users ...string) (Group, error) {
	//	Our return item
	retval := Group{}
	affectedUsers := []User{}

	//	First -- validate that the group exists
	err := store.systemdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(GetKey("Group", groupName))
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
		return retval, fmt.Errorf("Group does not exist")
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

	//	Get the group's new list of users from a merged (and deduped) list of:
	//	- The existing group users
	// 	- The list of users passed in
	allGroupUsers := append(retval.Users, users...)
	allUniqueGroupUsers := sort.StringSlice(allGroupUsers)

	sort.Sort(allUniqueGroupUsers)                // sort the data first
	n := set.Uniq(allUniqueGroupUsers)            // Uniq returns the size of the set
	allUniqueGroupUsers = allUniqueGroupUsers[:n] // trim the duplicate elements

	//	Then add each of users to both ...
	//	the usergroup
	retval.Users = allUniqueGroupUsers

	//	Serialize the group to JSON format
	encoded, err := json.Marshal(retval)
	if err != nil {
		return retval, fmt.Errorf("Problem serializing the data: %s", err)
	}

	//	Save it to the database:
	err = store.systemdb.Update(func(txn *badger.Txn) error {
		err := txn.Set(GetKey("Group", retval.Name), encoded)
		if err != nil {
			return err
		}

		//	Save each affected user to the database (as part of the same transaction):
		for _, affectedUser := range affectedUsers {

			//	and add the group to each user
			currentGroups := append(affectedUser.Groups, groupName)
			allUniqueCurrentGroups := sort.StringSlice(currentGroups)

			sort.Sort(allUniqueCurrentGroups)                    // sort the data first
			cn := set.Uniq(allUniqueCurrentGroups)               // Uniq returns the size of the set
			allUniqueCurrentGroups = allUniqueCurrentGroups[:cn] // trim the duplicate elements
			affectedUser.Groups = allUniqueCurrentGroups

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
		return retval, fmt.Errorf("Problem completing the group updates: %s", err)
	}

	//	Return our data:
	return retval, nil
}
