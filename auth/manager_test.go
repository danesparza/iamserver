package auth_test

import (
	"os"
	"testing"
	"time"

	"github.com/danesparza/iamserver/auth"
	"github.com/danesparza/iamserver/data"
	"github.com/danesparza/iamserver/policy"
)

func TestManager_DoPoliciesAllow_ValidRequest_Successful(t *testing.T) {

	//	Arrange
	mgr := &auth.Manager{}
	pols := map[string]data.Policy{
		"Regular user ship access": {
			Name:   "Regular user ship access",
			Effect: policy.Allow,
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"Find",
				"Open",
				"Embark",
				"Disembark",
			},
		},
		"Captain privledges": {
			Name:   "Captain privledges",
			Effect: policy.Allow,
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"Fly",
				"Navigate",
				"Curse",
			},
		},
		"Secret compartment access": {
			Name:   "Secret compartment access",
			Effect: policy.Allow,
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"AccessCompartments",
			},
		},
		"Healthcare access": {
			Name:   "Healthcare access",
			Effect: policy.Allow,
			Resources: []string{
				"Healthcare",
			},
			Actions: []string{
				"PresentHMOcard",
				"WaitToSeeDoc",
				"GetMedicalAdvice",
			},
		},
	}

	req := &data.Request{
		Action:   "Embark",
		Resource: "Serenity",
		User:     "malreynolds",
	}

	//	Act
	err := mgr.DoPoliciesAllow(req, pols)

	//	Assert
	if err != nil {
		t.Errorf("DoPoliciesAllow - should allow request, but got error: %v", err)
	}

}

func TestManager_DoPoliciesAllow_InvalidRequest_ReturnsError(t *testing.T) {

	//	Arrange
	mgr := &auth.Manager{}
	pols := map[string]data.Policy{
		"Regular user ship access": {
			Name:   "Regular user ship access",
			Effect: policy.Allow,
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"Find",
				"Open",
				"Embark",
				"Disembark",
			},
		},
		"Captain privledges": {
			Name:   "Captain privledges",
			Effect: policy.Allow,
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"Fly",
				"Navigate",
				"Curse",
			},
		},
		"Secret compartment access": {
			Name:   "Secret compartment access",
			Effect: policy.Allow,
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"AccessCompartments",
			},
		},
		"Healthcare access": {
			Name:   "Healthcare access",
			Effect: policy.Allow,
			Resources: []string{
				"Healthcare",
			},
			Actions: []string{
				"PresentHMOcard",
				"WaitToSeeDoc",
				"GetMedicalAdvice",
			},
		},
	}

	req := &data.Request{
		Action:   "Fire",
		Resource: "Serenity",
		User:     "malreynolds",
	}

	//	Act
	err := mgr.DoPoliciesAllow(req, pols)

	//	Assert
	if err == nil {
		t.Errorf("DoPoliciesAllow - should implicitly deny request, but did not get error")
	}

}

func TestManager_DoPoliciesAllow_ExplicitDeny_ReturnsError(t *testing.T) {

	//	Arrange
	mgr := &auth.Manager{}
	pols := map[string]data.Policy{
		"Regular user ship access": {
			Name:   "Regular user ship access",
			Effect: policy.Allow,
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"Find",
				"Open",
				"Embark",
				"Disembark",
			},
		},
		"Captain privledges": {
			Name:   "Captain privledges",
			Effect: policy.Allow,
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"Fly",
				"Navigate",
				"Curse",
			},
		},
		"Deny all ship access": {
			Name:   "Deny all ship access",
			Effect: policy.Deny, // Policy deny
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"<.*>", // Using a regex wildcard
			},
		},
		"Secret compartment access": {
			Name:   "Secret compartment access",
			Effect: policy.Allow,
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"AccessCompartments",
			},
		},
		"Healthcare access": {
			Name:   "Healthcare access",
			Effect: policy.Allow,
			Resources: []string{
				"Healthcare",
			},
			Actions: []string{
				"PresentHMOcard",
				"WaitToSeeDoc",
				"GetMedicalAdvice",
			},
		},
	}

	//	Act
	err1 := mgr.DoPoliciesAllow(
		&data.Request{
			Action:   "Open",
			Resource: "Serenity",
			User:     "malreynolds",
		}, pols)

	err2 := mgr.DoPoliciesAllow(
		&data.Request{
			Action:   "PresentHMOcard",
			Resource: "Healthcare",
			User:     "malreynolds",
		}, pols)

	//	Assert
	if err1 == nil {
		t.Errorf("DoPoliciesAllow - should explicitly deny request, but did not get error")
	}

	if err2 != nil {
		t.Errorf("DoPoliciesAllow - should allow request, but got error: %v", err2)
	}

}

func TestManager_IsUserRequestAuthorized_AuthorizedRequest_ReturnsTrue(t *testing.T) {
	//	Arrange
	systemdb, tokendb := getTestFiles()
	db, err := data.NewManager(systemdb, tokendb)
	if err != nil {
		t.Errorf("NewManager failed: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(systemdb)
		os.RemoveAll(tokendb)
	}()

	authMgr := auth.Manager{DBManager: db}

	contextUser := data.User{Name: "System"}

	//	The rules we want to setup and validate (based shamelessly on Firefly):
	//	The users have 'access the ship' role (which has 'regular user ship access' policy)
	//	The captain has the 'fly the ship' policy
	//	Some users are browncoats - that usergroup has the 'access hidden compartments' policy
	//	Some users are alliance - the alliance usergroup has the 'access healthcare' role

	//	** Setup everything...
	//	-- Add users
	db.AddUser(contextUser, data.User{Name: "malreynolds"}, "lassiter")
	db.AddUser(contextUser, data.User{Name: "zoewashburn"}, "warrior")
	db.AddUser(contextUser, data.User{Name: "wash"}, "stars")
	db.AddUser(contextUser, data.User{Name: "inaraserra"}, "guild")
	db.AddUser(contextUser, data.User{Name: "jaynecobb"}, "noreavers")
	db.AddUser(contextUser, data.User{Name: "kayleefrye"}, "shiny")
	db.AddUser(contextUser, data.User{Name: "simontam"}, "familyfirst")
	db.AddUser(contextUser, data.User{Name: "rivertam"}, "twobytwo")
	db.AddUser(contextUser, data.User{Name: "book"}, "shadowy")

	db.AddUser(contextUser, data.User{Name: "dobson"}, "alliancerulez")
	db.AddUser(contextUser, data.User{Name: "drcaron"}, "miranda")
	db.AddUser(contextUser, data.User{Name: "magistratehiggins"}, "mymoon")

	//	-- Add groups
	db.AddGroup(contextUser, "Browncoats", "")
	db.AddGroup(contextUser, "Alliance", "")

	//	-- Add users to groups
	_, err = db.AddUsersToGroup(contextUser, "Browncoats", "malreynolds", "zoewashburn", "wash", "inaraserra", "jaynecobb", "kayleefrye")
	if err != nil {
		t.Errorf("AddUsersToGroup - Should add users to 'Browncoats' group without an error, but got %s", err)
	}

	_, err = db.AddUsersToGroup(contextUser, "Alliance", "dobson", "drcaron", "magistratehiggins")
	if err != nil {
		t.Errorf("AddUsersToGroup - Should add users to 'Alliance' group without an error, but got %s", err)
	}

	//	-- Add roles
	db.AddRole(contextUser, "Ship access", "Can access the ship")
	db.AddRole(contextUser, "Healthcare access", "Can access alliance healthcare")

	//	-- Add resources & policies
	db.AddResource(contextUser, "Serenity", "The ship resource")
	db.AddResource(contextUser, "Healthcare", "Alliance healthcare resource")

	shipUser := data.Policy{
		Name:   "Regular user ship access",
		Effect: policy.Allow,
		Resources: []string{
			"Serenity",
		},
		Actions: []string{
			"Find",
			"Open",
			"Embark",
			"Disembark",
		},
	}
	db.AddPolicy(contextUser, shipUser)

	captainAccess := data.Policy{
		Name:   "Captain privledges",
		Effect: policy.Allow,
		Resources: []string{
			"Serenity",
		},
		Actions: []string{
			"Fly",
			"Navigate",
			"Curse",
		},
	}
	db.AddPolicy(contextUser, captainAccess)

	compartmentAccess := data.Policy{
		Name:   "Secret compartment access",
		Effect: policy.Allow,
		Resources: []string{
			"Serenity",
		},
		Actions: []string{
			"AccessCompartments",
		},
	}
	db.AddPolicy(contextUser, compartmentAccess)

	healthcareAccess := data.Policy{
		Name:   "Healthcare access",
		Effect: policy.Allow,
		Resources: []string{
			"Healthcare",
		},
		Actions: []string{
			"PresentHMOcard",
			"WaitToSeeDoc",
			"GetMedicalAdvice",
		},
	}
	db.AddPolicy(contextUser, healthcareAccess)

	//	-- Add policies to roles
	_, err = db.AttachPoliciesToRole(contextUser, "Ship access", shipUser.Name)
	if err != nil {
		t.Errorf("AttachPoliciesToRole - Should attach policy to role without an error, but got %s", err)
	}

	_, err = db.AttachPoliciesToRole(contextUser, "Healthcare access", healthcareAccess.Name)
	if err != nil {
		t.Errorf("AttachPoliciesToRole - Should attach policy to role without an error, but got %s", err)
	}

	//	** Now we can actually start assigning stuff ...

	//	-- Add role to all users
	allUsers, _ := db.GetAllUsers(contextUser)
	allUserNames := []string{}
	for _, user := range allUsers {
		allUserNames = append(allUserNames, user.Name)
	}
	db.AttachRoleToUsers(contextUser, "Ship access", allUserNames...)

	//	-- Add policy to captain
	db.AttachPolicyToUsers(contextUser, captainAccess.Name, "malreynolds")

	//	-- Add policy to browncoats usergroup
	db.AttachPolicyToGroups(contextUser, compartmentAccess.Name, "Browncoats")

	//  -- Add role to alliance users
	db.AttachRoleToGroups(contextUser, "Healthcare access", "Alliance")

	//	-- Create a request
	request := data.Request{
		User:     "malreynolds",
		Resource: "Serenity",
		Action:   "Fly",
	}

	start := time.Now() // Starting the stopwatch

	//	ACT
	isAuthorized := authMgr.IsUserRequestAuthorized(&request)

	stop := time.Now()                                            // Stopping the stopwatch
	elapsed := stop.Sub(start)                                    // Figuring out the time elapsed
	t.Logf("IsUserRequestAuthorized elapsed time: %v\n", elapsed) // Logging elapsed time

	//	ASSERT
	if isAuthorized == false {
		t.Errorf("IsUserRequestAuthorized - should be authorized, but returned 'false'")
	}

}

func TestManager_IsUserRequestAuthorized_NotAuthorized_ReturnsFalse(t *testing.T) {
	//	Arrange
	systemdb, tokendb := getTestFiles()
	db, err := data.NewManager(systemdb, tokendb)
	if err != nil {
		t.Errorf("NewManager failed: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(systemdb)
		os.RemoveAll(tokendb)
	}()

	authMgr := auth.Manager{DBManager: db}

	contextUser := data.User{Name: "System"}

	//	The rules we want to setup and validate (based shamelessly on Firefly):
	//	The users have 'access the ship' role (which has 'regular user ship access' policy)
	//	The captain has the 'fly the ship' policy
	//	Some users are browncoats - that usergroup has the 'access hidden compartments' policy
	//	Some users are alliance - the alliance usergroup has the 'access healthcare' role

	//	** Setup everything...
	//	-- Add users
	db.AddUser(contextUser, data.User{Name: "malreynolds"}, "lassiter")
	db.AddUser(contextUser, data.User{Name: "zoewashburn"}, "warrior")
	db.AddUser(contextUser, data.User{Name: "wash"}, "stars")
	db.AddUser(contextUser, data.User{Name: "inaraserra"}, "guild")
	db.AddUser(contextUser, data.User{Name: "jaynecobb"}, "noreavers")
	db.AddUser(contextUser, data.User{Name: "kayleefrye"}, "shiny")
	db.AddUser(contextUser, data.User{Name: "simontam"}, "familyfirst")
	db.AddUser(contextUser, data.User{Name: "rivertam"}, "twobytwo")
	db.AddUser(contextUser, data.User{Name: "book"}, "shadowy")

	db.AddUser(contextUser, data.User{Name: "dobson"}, "alliancerulez")
	db.AddUser(contextUser, data.User{Name: "drcaron"}, "miranda")
	db.AddUser(contextUser, data.User{Name: "magistratehiggins"}, "mymoon")

	//	-- Add groups
	db.AddGroup(contextUser, "Browncoats", "")
	db.AddGroup(contextUser, "Alliance", "")

	//	-- Add users to groups
	_, err = db.AddUsersToGroup(contextUser, "Browncoats", "malreynolds", "zoewashburn", "wash", "inaraserra", "jaynecobb", "kayleefrye")
	if err != nil {
		t.Errorf("AddUsersToGroup - Should add users to 'Browncoats' group without an error, but got %s", err)
	}

	_, err = db.AddUsersToGroup(contextUser, "Alliance", "dobson", "drcaron", "magistratehiggins")
	if err != nil {
		t.Errorf("AddUsersToGroup - Should add users to 'Alliance' group without an error, but got %s", err)
	}

	//	-- Add roles
	db.AddRole(contextUser, "Ship access", "Can access the ship")
	db.AddRole(contextUser, "Healthcare access", "Can access alliance healthcare")

	//	-- Add resources & policies
	db.AddResource(contextUser, "Serenity", "The ship resource")
	db.AddResource(contextUser, "Healthcare", "Alliance healthcare resource")

	shipUser := data.Policy{
		Name:   "Regular user ship access",
		Effect: policy.Allow,
		Resources: []string{
			"Serenity",
		},
		Actions: []string{
			"Find",
			"Open",
			"Embark",
			"Disembark",
		},
	}
	db.AddPolicy(contextUser, shipUser)

	captainAccess := data.Policy{
		Name:   "Captain privledges",
		Effect: policy.Allow,
		Resources: []string{
			"Serenity",
		},
		Actions: []string{
			"Fly",
			"Navigate",
			"Curse",
		},
	}
	db.AddPolicy(contextUser, captainAccess)

	compartmentAccess := data.Policy{
		Name:   "Secret compartment access",
		Effect: policy.Allow,
		Resources: []string{
			"Serenity",
		},
		Actions: []string{
			"AccessCompartments",
		},
	}
	db.AddPolicy(contextUser, compartmentAccess)

	healthcareAccess := data.Policy{
		Name:   "Healthcare access",
		Effect: policy.Allow,
		Resources: []string{
			"Healthcare",
		},
		Actions: []string{
			"PresentHMOcard",
			"WaitToSeeDoc",
			"GetMedicalAdvice",
		},
	}
	db.AddPolicy(contextUser, healthcareAccess)

	//	-- Add policies to roles
	_, err = db.AttachPoliciesToRole(contextUser, "Ship access", shipUser.Name)
	if err != nil {
		t.Errorf("AttachPoliciesToRole - Should attach policy to role without an error, but got %s", err)
	}

	_, err = db.AttachPoliciesToRole(contextUser, "Healthcare access", healthcareAccess.Name)
	if err != nil {
		t.Errorf("AttachPoliciesToRole - Should attach policy to role without an error, but got %s", err)
	}

	//	** Now we can actually start assigning stuff ...

	//	-- Add role to all users
	allUsers, _ := db.GetAllUsers(contextUser)
	allUserNames := []string{}
	for _, user := range allUsers {
		allUserNames = append(allUserNames, user.Name)
	}
	db.AttachRoleToUsers(contextUser, "Ship access", allUserNames...)

	//	-- Add policy to captain
	db.AttachPolicyToUsers(contextUser, captainAccess.Name, "malreynolds")

	//	-- Add policy to browncoats usergroup
	db.AttachPolicyToGroups(contextUser, compartmentAccess.Name, "Browncoats")

	//  -- Add role to alliance users
	db.AttachRoleToGroups(contextUser, "Healthcare access", "Alliance")

	//	-- Create a request
	request := data.Request{
		User:     "malreynolds",
		Resource: "Serenity",
		Action:   "Kiss_a_reaver",
	}

	start := time.Now() // Starting the stopwatch

	//	ACT
	isAuthorized := authMgr.IsUserRequestAuthorized(&request)

	stop := time.Now()                                            // Stopping the stopwatch
	elapsed := stop.Sub(start)                                    // Figuring out the time elapsed
	t.Logf("IsUserRequestAuthorized elapsed time: %v\n", elapsed) // Logging elapsed time

	//	ASSERT
	if isAuthorized == true {
		t.Errorf("IsUserRequestAuthorized - should NOT be authorized, but returned 'true'")
	}

}
