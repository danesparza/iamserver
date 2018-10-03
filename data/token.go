package data

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/danesparza/badger"
	"github.com/rs/xid"
)

// Token represents an auth token
type Token struct {
	ID      string    `json:"token"`
	User    string    `json:"user"`
	Created time.Time `json:"created"`
	Expires time.Time `json:"expires"`
}

// GetNewToken gets a token for the given user.  The token will have a TTL and expire automatically
func (store Manager) GetNewToken(user User, expiresafter time.Duration) (Token, error) {

	retval := Token{}

	//	Make sure the user exists first
	err := store.systemdb.View(func(txn *badger.Txn) error {
		_, err := txn.Get(GetKey("User", user.Name))
		return err
	})

	if err != nil {
		return retval, fmt.Errorf("User %s doesn't exist", user.Name)
	}

	//	Create our default return value
	newToken := Token{
		ID:      xid.New().String(), // Generate a new token
		User:    user.Name,
		Created: time.Now(),
		Expires: time.Now().Add(expiresafter),
	}

	//	Serialize to JSON format
	encoded, err := json.Marshal(newToken)
	if err != nil {
		return retval, fmt.Errorf("Problem serializing the token: %s", err)
	}

	//	Save it to the database:
	err = store.tokendb.Update(func(txn *badger.Txn) error {
		err := txn.SetWithTTL(GetKey("Token", newToken.ID), encoded, expiresafter)
		return err
	})

	//	If there was an error saving the data, report it:
	if err != nil {
		return retval, fmt.Errorf("Problem saving the token: %s", err)
	}

	//	Set our retval:
	retval = newToken

	//	Return the token
	return retval, nil
}

// GetTokenInfo returns token information for a given unexpired tokenID (or an error if it can't be found)
func (store Manager) GetTokenInfo(tokenID string) (Token, error) {

	retval := Token{}

	//	Get the token:
	err := store.tokendb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(GetKey("Token", tokenID))
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

	if err != nil {
		return retval, fmt.Errorf("Token %s doesn't exist", tokenID)
	}

	//	Return the token
	return retval, nil
}

// GetUserForToken returns user information for a given unexpired tokenID (or an error if token or user can't be found)
func (store Manager) GetUserForToken(tokenID string) (User, error) {

	retval := User{}
	token := Token{}

	//	First, see if we can get the token...
	err := store.tokendb.View(func(txn *badger.Txn) error {

		item, err := txn.Get(GetKey("Token", tokenID))
		if err != nil {
			return fmt.Errorf("Token doesn't exist: %s", tokenID)
		}
		val, err := item.Value()
		if err != nil {
			return err
		}

		if len(val) > 0 {
			//	Unmarshal data into our item
			if err := json.Unmarshal(val, &token); err != nil {
				return fmt.Errorf("Problem deserializing token %s: %s", tokenID, err)
			}
		}

		return nil
	})

	if err != nil {
		return retval, err
	}

	//	Next, see if we can get the user...
	err = store.systemdb.View(func(txn *badger.Txn) error {

		item, err := txn.Get(GetKey("User", token.User))
		if err != nil {
			return fmt.Errorf("User doesn't exist: %s", token.User)
		}
		val, err := item.Value()
		if err != nil {
			return err
		}

		if len(val) > 0 {
			//	Unmarshal data into our item
			if err := json.Unmarshal(val, &retval); err != nil {
				return fmt.Errorf("Problem deserializing user %s: %s", token.User, err)
			}
		}

		return nil
	})

	if err != nil {
		return retval, err
	}

	//	Return the user
	return retval, nil
}
