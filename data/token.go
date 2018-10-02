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
