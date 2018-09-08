package data

import (
	"fmt"
	"strings"

	"github.com/danesparza/badger"
)

// Manager is the data manager
type Manager struct {
	systemdb *badger.DB
	tokendb  *badger.DB
}

// NewManager creates a new instance of a Manager and returns it
func NewManager(systemdbpath, tokendbpath string) (*Manager, error) {
	retval := new(Manager)

	//	Open the systemDB
	sysopts := badger.DefaultOptions
	sysopts.Dir = systemdbpath
	sysopts.ValueDir = systemdbpath
	sysdb, err := badger.Open(sysopts)
	if err != nil {
		return retval, fmt.Errorf("Problem opening the systemDB: %s", err)
	}
	retval.systemdb = sysdb

	//	Open the tokenDB
	tokopts := badger.DefaultOptions
	tokopts.Dir = tokendbpath
	tokopts.ValueDir = tokendbpath
	tokdb, err := badger.Open(tokopts)
	if err != nil {
		return retval, fmt.Errorf("Problem opening the tokenDB: %s", err)
	}
	retval.tokendb = tokdb

	//	Return our Manager reference
	return retval, nil
}

// Close closes the data Manager
func (store Manager) Close() error {
	syserr := store.systemdb.Close()
	tokerr := store.tokendb.Close()

	if syserr != nil || tokerr != nil {
		return fmt.Errorf("An error occurred closing the manager.  Syserr: %s / Tokerr: %s", syserr, tokerr)
	}

	return nil
}

// GetKey returns a key to be used in the storage system
func GetKey(entityType string, keyPart ...string) []byte {
	allparts := []string{}
	allparts = append(allparts, entityType)
	allparts = append(allparts, keyPart...)
	return []byte(strings.Join(allparts, "_"))
}
