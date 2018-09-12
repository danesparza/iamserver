package data

import (
	"fmt"
	"strings"

	"github.com/danesparza/badger"
	"github.com/rs/xid"
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

// AuthSystemBootstrap is a 'run-once' operation to setup up the system initially
func (store Manager) AuthSystemBootstrap() (User, string, error) {
	adminUser := User{}

	//	Generate a password for the admin user
	adminPassword := xid.New().String()

	//	Create the admin user
	contextUser := User{Name: "System"}
	adminUser, err := store.AddUser(contextUser, User{Name: "admin"}, adminPassword)

	if err != nil {
		return adminUser, adminPassword, fmt.Errorf("Problem adding admin user: %s", err)
	}

	//	Create the Administrators group (add the admin user to the group)

	//	Create the system resource

	//	Create the initial system policies

	//	Create the System_Admin role (and add some of the system policies to that role)

	//	Attach the System_Admin role to the Administrators group

	//	Return everything:
	return adminUser, adminPassword, nil
}

// GetKey returns a key to be used in the storage system
func GetKey(entityType string, keyPart ...string) []byte {
	allparts := []string{}
	allparts = append(allparts, entityType)
	allparts = append(allparts, keyPart...)
	return []byte(strings.Join(allparts, "_"))
}
