package data

import (
	"fmt"
	"regexp"
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
	sysreqDeleteUser           = &Request{"System", "DeleteUser"}
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
	sysreqGetPolicy            = &Request{"System", "GetPolicy"}
	sysreqGetAllPolicies       = &Request{"System", "GetAllPolicies"}
	sysreqAttachPolicyToUsers  = &Request{"System", "AttachPolicyToUsers"}
	sysreqAttachPolicyToGroups = &Request{"System", "AttachPolicyToGroups"}
	sysreqGetPoliciesForUser   = &Request{"System", "GetPoliciesForUser"}
)

// SystemOverview represents the system overview data
type SystemOverview struct {
	UserCount     int
	GroupCount    int
	RoleCount     int
	PolicyCount   int
	ResourceCount int
}

// SearchResults represents search results
type SearchResults struct {
	Users     []string
	Groups    []string
	Roles     []string
	Policies  []string
	Resources []string
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
		sysreqGetUser.Action,
		sysreqGetAllUsers.Action,
		sysreqAddGroup.Action,
		sysreqGetGroup.Action,
		sysreqGetAllGroups.Action,
		sysreqAddUsersToGroup.Action,
		sysreqAddResource.Action,
		sysreqGetResource.Action,
		sysreqGetAllResources.Action,
		sysreqAddActionToResource.Action,
		sysreqAddRole.Action,
		sysreqGetRole.Action,
		sysreqGetAllRoles.Action,
		sysreqAttachPoliciesToRole.Action,
		sysreqAttachRoleToUsers.Action,
		sysreqAttachRoleToGroups.Action,
		sysreqAddPolicy.Action,
		sysreqGetPolicy.Action,
		sysreqGetAllPolicies.Action,
		sysreqAttachPolicyToUsers.Action,
		sysreqAttachPolicyToGroups.Action,
		sysreqGetPoliciesForUser.Action,
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

// GetOverview gets a system overview of counts in the system
func (store Manager) GetOverview(context User) (SystemOverview, error) {
	retval := SystemOverview{}

	//	Group count
	groupCount := 0
	if store.IsUserRequestAuthorized(context, sysreqGetAllGroups) {
		err := store.systemdb.View(func(txn *badger.Txn) error {

			//	Get an iterator - but don't prefetch values (key-only search)
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false
			it := txn.NewIterator(opts)
			defer it.Close()

			//	Set our prefix
			prefix := GetKey("Group")

			//	Iterate over our values:
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				//	Increment the group count
				groupCount++
			}
			return nil
		})

		//	If there was an error, report it:
		if err != nil {
			return retval, fmt.Errorf("Problem getting the count of groups: %s", err)
		}
	}

	//	User count
	userCount := 0
	if store.IsUserRequestAuthorized(context, sysreqGetAllUsers) {
		err := store.systemdb.View(func(txn *badger.Txn) error {

			//	Get an iterator - but don't prefetch values (key-only search)
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false
			it := txn.NewIterator(opts)
			defer it.Close()

			//	Set our prefix
			prefix := GetKey("User")

			//	Iterate over our values:
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				//	Increment the user count
				userCount++
			}
			return nil
		})

		//	If there was an error, report it:
		if err != nil {
			return retval, fmt.Errorf("Problem getting the count of users: %s", err)
		}
	}

	//	Role count
	roleCount := 0
	if store.IsUserRequestAuthorized(context, sysreqGetAllRoles) {
		err := store.systemdb.View(func(txn *badger.Txn) error {

			//	Get an iterator - but don't prefetch values (key-only search)
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false
			it := txn.NewIterator(opts)
			defer it.Close()

			//	Set our prefix
			prefix := GetKey("Role")

			//	Iterate over our values:
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				//	Increment the role count
				roleCount++
			}
			return nil
		})

		//	If there was an error, report it:
		if err != nil {
			return retval, fmt.Errorf("Problem getting the count of roles: %s", err)
		}
	}

	//	Policy count
	policyCount := 0
	if store.IsUserRequestAuthorized(context, sysreqGetAllPolicies) {
		err := store.systemdb.View(func(txn *badger.Txn) error {

			//	Get an iterator - but don't prefetch values (key-only search)
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false
			it := txn.NewIterator(opts)
			defer it.Close()

			//	Set our prefix
			prefix := GetKey("Policy")

			//	Iterate over our values:
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				//	Increment the policy count
				policyCount++
			}
			return nil
		})

		//	If there was an error, report it:
		if err != nil {
			return retval, fmt.Errorf("Problem getting the count of policies: %s", err)
		}
	}

	//	Resource count
	resourceCount := 0
	if store.IsUserRequestAuthorized(context, sysreqGetAllResources) {
		err := store.systemdb.View(func(txn *badger.Txn) error {

			//	Get an iterator - but don't prefetch values (key-only search)
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false
			it := txn.NewIterator(opts)
			defer it.Close()

			//	Set our prefix
			prefix := GetKey("Resource")

			//	Iterate over our values:
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				//	Increment the resource count
				resourceCount++
			}
			return nil
		})

		//	If there was an error, report it:
		if err != nil {
			return retval, fmt.Errorf("Problem getting the count of resources: %s", err)
		}
	}

	//	Set counts:
	retval = SystemOverview{
		GroupCount:    groupCount,
		UserCount:     userCount,
		RoleCount:     roleCount,
		PolicyCount:   policyCount,
		ResourceCount: resourceCount,
	}

	//	Return our information:
	return retval, nil
}

// Search gets items that match the searchExpression
func (store Manager) Search(context User, searchExpression string) (SearchResults, error) {
	retval := SearchResults{}

	//	Make sure the regexp is a case insensitive search:
	searchExpression = "(?i)" + searchExpression

	//	Compile the search expression:
	r, err := regexp.Compile(searchExpression)
	if err != nil {
		return retval, fmt.Errorf("Problem with search expression: %s", err)
	}

	//	Groups found
	groups := []string{}
	if store.IsUserRequestAuthorized(context, sysreqGetAllGroups) {
		err := store.systemdb.View(func(txn *badger.Txn) error {

			//	Get an iterator
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false
			it := txn.NewIterator(opts)
			defer it.Close()

			//	Set our prefix
			prefix := GetKey("Group")

			//	Iterate over our values:
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				//	See if the search expression matches
				currentItem := string(it.Item().Key())
				if r.MatchString(currentItem) {
					groups = append(groups, currentItem[len(prefix)+1:])
				}
			}
			return nil
		})

		//	If there was an error, report it:
		if err != nil {
			return retval, fmt.Errorf("Problem searching through groups: %s", err)
		}
	}

	//	Users found
	users := []string{}
	if store.IsUserRequestAuthorized(context, sysreqGetAllUsers) {
		err := store.systemdb.View(func(txn *badger.Txn) error {

			//	Get an iterator
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false
			it := txn.NewIterator(opts)
			defer it.Close()

			//	Set our prefix
			prefix := GetKey("User")

			//	Iterate over our values:
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				//	See if the search expression matches
				currentItem := string(it.Item().Key())
				if r.MatchString(currentItem) {
					users = append(users, currentItem[len(prefix)+1:])
				}
			}
			return nil
		})

		//	If there was an error, report it:
		if err != nil {
			return retval, fmt.Errorf("Problem searching through users: %s", err)
		}
	}

	//	Roles found
	roles := []string{}
	if store.IsUserRequestAuthorized(context, sysreqGetAllRoles) {
		err := store.systemdb.View(func(txn *badger.Txn) error {

			//	Get an iterator
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false
			it := txn.NewIterator(opts)
			defer it.Close()

			//	Set our prefix
			prefix := GetKey("Role")

			//	Iterate over our values:
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				//	See if the search expression matches
				currentItem := string(it.Item().Key())
				if r.MatchString(currentItem) {
					roles = append(roles, currentItem[len(prefix)+1:])
				}
			}
			return nil
		})

		//	If there was an error, report it:
		if err != nil {
			return retval, fmt.Errorf("Problem searching through roles: %s", err)
		}
	}

	//	Policies found
	policies := []string{}
	if store.IsUserRequestAuthorized(context, sysreqGetAllPolicies) {
		err := store.systemdb.View(func(txn *badger.Txn) error {

			//	Get an iterator
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false
			it := txn.NewIterator(opts)
			defer it.Close()

			//	Set our prefix
			prefix := GetKey("Policy")

			//	Iterate over our values:
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				//	See if the search expression matches
				currentItem := string(it.Item().Key())
				if r.MatchString(currentItem) {
					policies = append(policies, currentItem[len(prefix)+1:])
				}
			}
			return nil
		})

		//	If there was an error, report it:
		if err != nil {
			return retval, fmt.Errorf("Problem searching through policies: %s", err)
		}
	}

	//	Resources found
	resources := []string{}
	if store.IsUserRequestAuthorized(context, sysreqGetAllResources) {
		err := store.systemdb.View(func(txn *badger.Txn) error {

			//	Get an iterator
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false
			it := txn.NewIterator(opts)
			defer it.Close()

			//	Set our prefix
			prefix := GetKey("Resource")

			//	Iterate over our values:
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				//	See if the search expression matches
				currentItem := string(it.Item().Key())
				if r.MatchString(currentItem) {
					resources = append(resources, currentItem[len(prefix)+1:])
				}
			}
			return nil
		})

		//	If there was an error, report it:
		if err != nil {
			return retval, fmt.Errorf("Problem searching through resources: %s", err)
		}
	}

	//	Set counts:
	retval = SearchResults{
		Groups:    groups,
		Users:     users,
		Roles:     roles,
		Policies:  policies,
		Resources: resources,
	}

	//	Return our information:
	return retval, nil
}

// GetKey returns a key to be used in the storage system
func GetKey(entityType string, keyPart ...string) []byte {
	allparts := []string{}
	allparts = append(allparts, entityType)
	allparts = append(allparts, keyPart...)
	return []byte(strings.Join(allparts, ":"))
}
