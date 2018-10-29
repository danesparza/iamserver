package data

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/danesparza/badger"
	"golang.org/x/crypto/bcrypt"
	null "gopkg.in/guregu/null.v3"
	"gopkg.in/guregu/null.v3/zero"
)

// User represents a user in the system.  Users
// are associated with resources and roles within those applications/resources/services.
// They can be created/updated/deleted.  If they are deleted, eventually
// they will be removed from the system.  The admin user can only be disabled, not deleted
type User struct {
	Name        string      `json:"name"`
	Enabled     bool        `json:"enabled"`
	Description string      `json:"description"`
	SecretHash  string      `json:"secrethash"`
	TOTPEnabled bool        `json:"totpenabled"`
	TOTPSecret  string      `json:"totpsecret"`
	Created     time.Time   `json:"created"`
	CreatedBy   string      `json:"created_by"`
	Updated     time.Time   `json:"updated"`
	UpdatedBy   string      `json:"updated_by"`
	Deleted     zero.Time   `json:"deleted"`
	DeletedBy   null.String `json:"deleted_by"`
	Groups      []string    `json:"groups"`
	Policies    []string    `json:"policies"`
	Roles       []string    `json:"roles"`
}

// AddUser adds a user to the system
func (store Manager) AddUser(context User, user User, userPassword string) (User, error) {
	//	Our return item
	retval := User{}

	//	Security check:  Are we authorized to perform this action?
	if !store.IsUserRequestAuthorized(context, sysreqAddUser) {
		return retval, fmt.Errorf("User %s is not authorized to perform the action", context.Name)
	}

	//	First -- does the user exist already?
	err := store.systemdb.View(func(txn *badger.Txn) error {
		_, err := txn.Get(GetKey("User", user.Name))
		return err
	})

	//	If we didn't get an error, we have a problem:
	if err == nil {
		return retval, fmt.Errorf("User already exists")
	}

	//	Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
	if err != nil {
		return retval, fmt.Errorf("Problem hashing user password: %s", err)
	}

	//	Update the secret field:
	user.SecretHash = string(hashedPassword)

	//	Make sure it's initially set to 'enabled':
	user.Enabled = true

	//	Make sure (when adding a new user) groups/policies/roles are empty:
	user.Groups = []string{}
	user.Policies = []string{}
	user.Roles = []string{}

	//	Update the created / updated fields:
	user.Created = time.Now()
	user.Updated = time.Now()
	user.CreatedBy = context.Name
	user.UpdatedBy = context.Name

	//	Serialize to JSON format
	encoded, err := json.Marshal(user)
	if err != nil {
		return retval, fmt.Errorf("Problem serializing the data: %s", err)
	}

	//	Save it to the database:
	err = store.systemdb.Update(func(txn *badger.Txn) error {
		err := txn.Set(GetKey("User", user.Name), encoded)
		return err
	})

	//	If there was an error saving the data, report it:
	if err != nil {
		return retval, fmt.Errorf("Problem saving the data: %s", err)
	}

	//	Set our retval:
	retval = user

	//	Return our data:
	return retval, nil
}

// GetUser gets a user from the system
func (store Manager) GetUser(context User, userName string) (User, error) {
	//	Our return item
	retval := User{}

	//	Security check:  Are we authorized to perform this action?
	if !store.IsUserRequestAuthorized(context, sysreqGetUser) {
		return retval, fmt.Errorf("User %s is not authorized to perform the action", context.Name)
	}

	err := store.systemdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(GetKey("User", userName))
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

// DeleteUser adds a user to the system
func (store Manager) DeleteUser(context User, user User, userPassword string) (User, error) {
	//	Our return item
	retval := User{}

	//	Security check:  Are we authorized to perform this action?
	if !store.IsUserRequestAuthorized(context, sysreqDeleteUser) {
		return retval, fmt.Errorf("User %s is not authorized to perform the action", context.Name)
	}

	//	First -- does the user exist already?
	err := store.systemdb.View(func(txn *badger.Txn) error {
		_, err := txn.Get(GetKey("User", user.Name))
		return err
	})

	//	If got an error, we have a problem:
	if err != nil {
		return retval, fmt.Errorf("User does not exist")
	}

	//	Make sure it's initially set to 'enabled':
	user.Enabled = false

	//	Remove the user from groups / roles / policies

	//	Reset the groups / roles / policies collections:
	user.Groups = []string{}
	user.Roles = []string{}
	user.Policies = []string{}

	//	Update the updated / deleted fields:
	user.Deleted = zero.TimeFrom(time.Now())
	user.Updated = time.Now()
	user.DeletedBy = null.StringFrom(context.Name)
	user.UpdatedBy = context.Name

	//	Serialize to JSON format
	encoded, err := json.Marshal(user)
	if err != nil {
		return retval, fmt.Errorf("Problem serializing the data: %s", err)
	}

	//	Save it to the database with a TTL:
	err = store.systemdb.Update(func(txn *badger.Txn) error {
		err := txn.SetWithTTL(GetKey("User", user.Name), encoded, 168*time.Hour) // Set with an expire time of 1 week
		return err
	})

	//	If there was an error saving the data, report it:
	if err != nil {
		return retval, fmt.Errorf("Problem saving the data: %s", err)
	}

	//	Set our retval:
	retval = user

	//	Return our data:
	return retval, nil
}

// GetAllUsers gets all users in the system
func (store Manager) GetAllUsers(context User) ([]User, error) {
	//	Our return item
	retval := []User{}

	//	Security check:  Are we authorized to perform this action?
	if !store.IsUserRequestAuthorized(context, sysreqGetAllUsers) {
		return retval, fmt.Errorf("User %s is not authorized to perform the action", context.Name)
	}

	err := store.systemdb.View(func(txn *badger.Txn) error {

		//	Get an iterator
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		//	Set our prefix
		prefix := GetKey("User")

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
				item := User{}

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

// GetUserWithCredentials gets a user given a set of credentials
func (store Manager) GetUserWithCredentials(name, secret string) (User, error) {
	retUser := User{}
	tmpUser := User{}

	err := store.systemdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(GetKey("User", name))
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}

		if len(val) > 0 {
			//	Unmarshal data into our item
			if err := json.Unmarshal(val, &tmpUser); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return retUser, fmt.Errorf("The user was not found or the password was incorrect")
	}

	// Compare the given password with the hash
	err = bcrypt.CompareHashAndPassword([]byte(tmpUser.SecretHash), []byte(secret))
	if err != nil { // nil means it is a match
		return retUser, fmt.Errorf("The user was not found or the password was incorrect")
	}

	//	If everything checks out, return the user:
	retUser = tmpUser

	//	Return what we found:
	return retUser, nil
}
