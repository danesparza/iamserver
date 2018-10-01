package data

import (
	"fmt"
	"strings"

	"github.com/danesparza/badger"
	"github.com/danesparza/iamserver/policy"
	"github.com/rs/xid"
)

// Manager is the data manager
type Manager struct {
	systemdb *badger.DB
	tokendb  *badger.DB
	Matcher  matcher
}

var (
	// SystemUser represents the system user
	SystemUser = User{Name: "System"}

	sysreqAddUser              = &Request{"System", "AddUser"}
	sysreqGetUser              = &Request{"System", "GetUser"}
	sysreqGetAllUsers          = &Request{"System", "GetAllUsers"}
	sysreqAddGroup             = &Request{"System", "AddGroup"}
	sysreqGetGroup             = &Request{"System", "GetGroup"}
	sysreqGetAllGroups         = &Request{"System", "GetAllGroups"}
	sysreqAddUsersToGroup      = &Request{"System", "AddUsersToGroup"}
	sysreqAddResource          = &Request{"System", "AddResource"}
	sysreqGetResource          = &Request{"System", "GetResource"}
	sysreqGetAllResources      = &Request{"System", "GetAllResources"}
	sysreqAddActionToResource  = &Request{"System", "AddActionToResource"}
	sysreqAddRole              = &Request{"System", "AddRole"}
	sysreqGetRole              = &Request{"System", "GetRole"}
	sysreqGetAllRoles          = &Request{"System", "GetAllRoles"}
	sysreqAttachPoliciesToRole = &Request{"System", "AttachPoliciesToRole"}
	sysreqAttachRoleToUsers    = &Request{"System", "AttachRoleToUsers"}
	sysreqAttachRoleToGroups   = &Request{"System", "AttachRoleToGroups"}
	sysreqAddPolicy            = &Request{"System", "AddPolicy"}
	sysreqAttachPolicyToUsers  = &Request{"System", "AttachPolicyToUsers"}
	sysreqAttachPolicyToGroups = &Request{"System", "AttachPolicyToGroups"}
	sysreqGetPoliciesForUser   = &Request{"System", "GetPoliciesForUser"}
)

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

// SystemBootstrap is a 'run-once' operation to setup up the system initially
func (store Manager) SystemBootstrap() (User, string, error) {
	adminUser := User{}
	contextUser := User{Name: "System"}

	//	Generate a password for the admin user
	adminPassword := xid.New().String()

	//	Create the admin user
	adminUser, err := store.AddUser(contextUser, User{Name: "admin", Description: "System administrator"}, adminPassword)
	if err != nil {
		return adminUser, adminPassword, fmt.Errorf("Problem adding admin user: %s", err)
	}

	//	Create the Administrators group (add the admin user to the group)
	adminGroup, err := store.AddGroup(contextUser, "Administrators", "Users who can fully administer the system")
	if err != nil {
		return adminUser, adminPassword, fmt.Errorf("Problem creating the Administrators group: %s", err)
	}

	_, err = store.AddUsersToGroup(contextUser, adminGroup.Name, adminUser.Name)
	if err != nil {
		return adminUser, adminPassword, fmt.Errorf("Problem adding the admin user to the Administrators group: %s", err)
	}

	//	Create the system resource
	_, err = store.AddResource(contextUser, "System", "The system resource")
	if err != nil {
		return adminUser, adminPassword, fmt.Errorf("Problem creating the system resource: %s", err)
	}

	//	Add system actions
	store.AddActionToResource(contextUser, "System",
		sysreqAddUser.Action,
		"EditUser",
		"GetUser",
		"ListUsers",
		"CreateGroup",
		"EditGroup",
		"GetGroup",
		"ListGroups",
		"CreatePolicy",
		"EditPolicy",
		"GetPolicy",
		"ListPolicies",
		"CreateRole",
		"EditRole",
		"GetRole",
		"ListRoles",
		"CreateResource",
		"EditResource",
		"GetResource",
		"ListResources",
		"CreateAction",
		"EditAction",
	)

	//	Create the initial system policies
	adminEverything := Policy{
		Name:   "Administer everything",
		Effect: policy.Allow,
		Resources: []string{
			"<.*>", // All resources
		},
		Actions: []string{
			"<.*>", // All actions
		},
	}
	_, err = store.AddPolicy(contextUser, adminEverything)
	if err != nil {
		return adminUser, adminPassword, fmt.Errorf("Problem creating the 'administer everything' policy: %s", err)
	}

	//	Create the sys_admin role (and add some of the system policies to that role)
	sysAdmin, err := store.AddRole(contextUser, "sys_admin", "System administrator role")
	if err != nil {
		return adminUser, adminPassword, fmt.Errorf("Problem creating the 'sys_admin' role: %s", err)
	}
	_, err = store.AttachPoliciesToRole(contextUser, sysAdmin.Name, adminEverything.Name)
	if err != nil {
		return adminUser, adminPassword, fmt.Errorf("Problem adding policies to the 'sys_admin' role: %s", err)
	}

	//	Attach the sys_admin role to the Administrators group
	_, err = store.AttachRoleToGroups(contextUser, sysAdmin.Name, adminGroup.Name)
	if err != nil {
		return adminUser, adminPassword, fmt.Errorf("Problem attaching 'sys_admin' role to the 'Administrators' group: %s", err)
	}

	//	Get the updated admin user:
	adminUser, err = store.GetUser(contextUser, adminUser.Name)
	if err != nil {
		return adminUser, adminPassword, fmt.Errorf("Problem getting the updated admin user: %s", err)
	}

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

// IsUserAuthorized validates the request
func (store Manager) IsUserAuthorized(user User, request Request) bool {

	return true
}
