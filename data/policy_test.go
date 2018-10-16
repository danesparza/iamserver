package data_test

import (
	"os"
	"testing"
	"time"

	"github.com/danesparza/iamserver/data"
	"github.com/danesparza/iamserver/policy"
)

func TestPolicy_AddPolicy_ValidPolicy_Successful(t *testing.T) {

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

	contextUser := data.User{Name: "System"}
	testPolicy := data.Policy{
		Name:   "UnitTest1",
		Effect: policy.Allow,
		Resources: []string{
			"Someresource",
		},
		Actions: []string{
			"Someaction",
		},
	}
	_, err = db.AddResource(contextUser, "Someresource", "")

	//	Act
	newPolicy, err := db.AddPolicy(contextUser, testPolicy)

	//	Assert
	if err != nil {
		t.Errorf("AddPolicy - Should add policy without error, but got: %s", err)
	}

	if newPolicy.Created.IsZero() || newPolicy.Updated.IsZero() {
		t.Errorf("AddPolicy failed: Should have set an item with the correct datetime: %+v", newPolicy)
	}

	if newPolicy.CreatedBy != contextUser.Name {
		t.Errorf("AddPolicy failed: Should have set an item with the correct 'created by' user: %+v", newPolicy)
	}

	if newPolicy.UpdatedBy != contextUser.Name {
		t.Errorf("AddPolicy failed: Should have set an item with the correct 'updated by' user: %+v", newPolicy)
	}
}

func TestPolicy_GetPolicy_ValidPolicy_Successful(t *testing.T) {

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

	contextUser := data.User{Name: "System"}
	testPolicy := data.Policy{
		Name:   "UnitTest1",
		Effect: policy.Allow,
		Resources: []string{
			"Someresource",
		},
		Actions: []string{
			"Someaction",
		},
	}
	_, err = db.AddResource(contextUser, "Someresource", "")

	//	Act
	newPolicy, err := db.AddPolicy(contextUser, testPolicy)

	//	Assert
	if err != nil {
		t.Errorf("AddPolicy - Should add policy without error, but got: %s", err)
	}

	newPolicy2, err := db.GetPolicy(contextUser, testPolicy.Name)

	if newPolicy2.Resources[0] != newPolicy.Resources[0] {
		t.Errorf("GetPolicy failed: Should have gotten an item with the correct 'resources': %+v", newPolicy)
	}

	if newPolicy2.Actions[0] != newPolicy.Actions[0] {
		t.Errorf("GetPolicy failed: Should have gotten an item with the correct 'actions': %+v", newPolicy)
	}
}

func TestPolicy_AddPolicy_AlreadyExists_ReturnsError(t *testing.T) {

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

	contextUser := data.User{Name: "System"}
	testPolicy := data.Policy{
		Name:   "UnitTest1",
		Effect: policy.Allow,
		Resources: []string{
			"Someresource",
		},
		Actions: []string{
			"Someaction",
		},
	}
	_, err = db.AddResource(contextUser, "Someresource", "")

	//	Act
	_, err = db.AddPolicy(contextUser, testPolicy)
	if err != nil {
		t.Errorf("AddPolicy - Should add policy without error, but got: %s", err)
	}
	_, err = db.AddPolicy(contextUser, testPolicy)

	//	Assert
	if err == nil {
		t.Errorf("AddPolicy - Should not add duplicate policy without error")
	}

}

func TestPolicy_AddPolicy_InvalidEffect_ReturnsError(t *testing.T) {

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

	contextUser := data.User{Name: "System"}
	testPolicy := data.Policy{
		Name:   "UnitTest1",
		Effect: "someweirdeffect",
		Resources: []string{
			"Someresource",
		},
		Actions: []string{
			"Someaction",
		},
	}
	_, err = db.AddResource(contextUser, "Someresource", "")

	//	Act
	_, err = db.AddPolicy(contextUser, testPolicy)

	//	Assert
	if err == nil {
		t.Errorf("AddPolicy - Should not add policy with invalid effect")
	}

}

func TestPolicy_AttachPoliciesToUser_PolicyDoesntExist_ReturnsError(t *testing.T) {

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

	contextUser := data.User{Name: "System"}

	//	Act

	//	Add some users
	db.AddUser(contextUser, data.User{Name: "Unittestuser1"}, "testpass")
	db.AddUser(contextUser, data.User{Name: "Unittestuser2"}, "testpass")
	db.AddUser(contextUser, data.User{Name: "Unittestuser3"}, "testpass")
	db.AddUser(contextUser, data.User{Name: "Unittestuser4"}, "testpass")

	//	Attempt to attach policies that don't exist yet
	retpolicy, err := db.AttachPolicyToUsers(contextUser, "Bad policy 1", "Unittestuser1", "Unittestuser2", "Unittestuser3")

	// Sanity check the error
	// t.Logf("AttachPolicyToUsers error: %s", err)

	if len(retpolicy.Users) > 0 {
		t.Errorf("AttachPolicyToUsers - Should not have attached policies that don't exist.")
	}

	//	Assert
	if err == nil {
		t.Errorf("AttachPolicyToUsers - Should throw error attempting to attach policies that don't exist but didn't get an error")
	}
}

func TestPolicy_AttachPoliciesToGroup_PolicyDoesntExist_ReturnsError(t *testing.T) {

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

	contextUser := data.User{Name: "System"}

	//	Act

	//	Add some groups
	db.AddGroup(contextUser, "Unittestgroup1", "")
	db.AddGroup(contextUser, "Unittestgroup2", "")
	db.AddGroup(contextUser, "Unittestgroup3", "")
	db.AddGroup(contextUser, "Unittestgroup4", "")

	//	Attempt to attach policies that don't exist yet
	retpolicy, err := db.AttachPolicyToGroups(contextUser, "Bad policy 1", "Unittestgroup1", "Unittestgroup2", "Unittestgroup3")

	// Sanity check the error
	// t.Logf("AttachPolicyToGroups error: %s", err)

	if len(retpolicy.Groups) > 0 {
		t.Errorf("AttachPolicyToGroups - Should not have attached policies that don't exist.")
	}

	//	Assert
	if err == nil {
		t.Errorf("AttachPolicyToGroups - Should throw error attempting to attach policies that don't exist but didn't get an error")
	}

}

func TestPolicy_AttachPoliciesToUser_UserDoesntExist_ReturnsError(t *testing.T) {

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

	contextUser := data.User{Name: "System"}

	//	Add a resource
	db.AddResource(contextUser, "Someresource", "")

	//	Add a policy
	testPolicy := data.Policy{
		Name:   "UnitTest1",
		Effect: policy.Allow,
		Resources: []string{
			"Someresource",
		},
		Actions: []string{
			"Someaction",
		},
	}

	newPolicy, err := db.AddPolicy(contextUser, testPolicy)

	//	Act

	//	Attempt to attach policies to users that don't exist yet
	retpolicy, err := db.AttachPolicyToUsers(contextUser, newPolicy.Name, "Unittestuser1", "Unittestuser2", "Unittestuser3")

	// Sanity check the error
	// t.Logf("AttachPolicyToUsers error: %s", err)

	if len(retpolicy.Users) > 0 {
		t.Errorf("AttachPolicyToUsers - Should not have attached policies to users that don't exist.")
	}

	//	Assert
	if err == nil {
		t.Errorf("AttachPolicyToUsers - Should throw error attempting to attach policies to users that don't exist but didn't get an error")
	}
}

func TestPolicy_AttachPoliciesToGroup_GroupDoesntExist_ReturnsError(t *testing.T) {

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

	contextUser := data.User{Name: "System"}

	//	Add a resource
	db.AddResource(contextUser, "Someresource", "")

	//	Add a policy
	testPolicy := data.Policy{
		Name:   "UnitTest1",
		Effect: policy.Allow,
		Resources: []string{
			"Someresource",
		},
		Actions: []string{
			"Someaction",
		},
	}

	newPolicy, err := db.AddPolicy(contextUser, testPolicy)

	//	Act

	//	Attempt to attach policies to groups that don't exist yet
	retpolicy, err := db.AttachPolicyToGroups(contextUser, newPolicy.Name, "Unittestgroup1", "Unittestgroup2", "Unittestgroup3")

	// Sanity check the error
	// t.Logf("AttachPolicyToGroups error: %s", err)

	if len(retpolicy.Groups) > 0 {
		t.Errorf("AttachPolicyToGroups - Should not have attached policies to groups that don't exist.")
	}

	//	Assert
	if err == nil {
		t.Errorf("AttachPolicyToGroups - Should throw error attempting to attach policies to groups that don't exist but didn't get an error")
	}

}

func TestPolicy_AttachPoliciesToUser_ValidParams_ReturnsPolicy(t *testing.T) {

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

	contextUser := data.User{Name: "System"}

	//	Act

	//	Add some users
	db.AddUser(contextUser, data.User{Name: "Unittestuser1"}, "testpass")
	db.AddUser(contextUser, data.User{Name: "Unittestuser2"}, "testpass")
	db.AddUser(contextUser, data.User{Name: "Unittestuser3"}, "testpass")
	db.AddUser(contextUser, data.User{Name: "Unittestuser4"}, "testpass")

	//	Add a resource
	db.AddResource(contextUser, "Someresource", "")

	//	Add a policy
	testPolicy := data.Policy{
		Name:   "UnitTest1",
		Effect: policy.Allow,
		Resources: []string{
			"Someresource",
		},
		Actions: []string{
			"Someaction",
		},
	}

	newPolicy, err := db.AddPolicy(contextUser, testPolicy)

	//	Attempt to attach the policy to the users
	retpolicy, err := db.AttachPolicyToUsers(contextUser, newPolicy.Name, "Unittestuser1", "Unittestuser2", "Unittestuser3")

	//	Assert
	if err != nil {
		t.Errorf("AttachPolicyToUsers - Should attach policy without an error, but got %s", err)
	}

	if len(retpolicy.Users) != 3 {
		t.Errorf("AttachPolicyToUsers - Should have attached policy to 3 users")
	}

	//	Sanity check the list of users:
	// t.Logf("Updated policy -- %+v", retpolicy)

	//	Sanity check that the users have the new policy now:
	user1, _ := db.GetUser(contextUser, "Unittestuser1")

	// t.Logf("Updated user -- %+v", user1)

	if len(user1.Policies) == 0 {
		t.Errorf("AttachPolicyToUsers - Should have attached policy to Unittestuser1, but policy is not attached to the user")
	}
}

func TestPolicy_AttachPoliciesToGroup_ValidParams_ReturnsPolicy(t *testing.T) {

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

	contextUser := data.User{Name: "System"}

	//	Act

	//	Add some groups
	db.AddGroup(contextUser, "Unittestgroup1", "")
	db.AddGroup(contextUser, "Unittestgroup2", "")
	db.AddGroup(contextUser, "Unittestgroup3", "")
	db.AddGroup(contextUser, "Unittestgroup4", "")

	//	Add a resource
	db.AddResource(contextUser, "Someresource", "")

	//	Add a policy
	testPolicy := data.Policy{
		Name:   "UnitTest1",
		Effect: policy.Allow,
		Resources: []string{
			"Someresource",
		},
		Actions: []string{
			"Someaction",
		},
	}

	newPolicy, err := db.AddPolicy(contextUser, testPolicy)

	//	Attempt to attach the policy to the groups
	retpolicy, err := db.AttachPolicyToGroups(contextUser, newPolicy.Name, "Unittestgroup1", "Unittestgroup2", "Unittestgroup3")

	//	Assert
	if err != nil {
		t.Errorf("AttachPolicyToGroups - Should attach policy without an error, but got %s", err)
	}

	if len(retpolicy.Groups) != 3 {
		t.Errorf("AttachPolicyToGroups - Should have attached policy to 3 groups")
	}

	//	Sanity check the list of groups:
	// t.Logf("Updated policy -- %+v", retpolicy)

	//	Sanity check that the groups have the new policy now:
	group1, _ := db.GetGroup(contextUser, "Unittestgroup1")

	// t.Logf("Updated group -- %+v", group1)

	if len(group1.Policies) == 0 {
		t.Errorf("AttachPolicyToGroups - Should have attached policy to Unittestgroup1, but policy is not attached")
	}
}

func TestPolicy_GetPoliciesForUser_ValidParams_ReturnsPolicies(t *testing.T) {
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

	//	ACT
	start := time.Now() // Starting the stopwatch

	malPolicies, err := db.GetPoliciesForUser(contextUser, "malreynolds")

	stop := time.Now()                                       // Stopping the stopwatch
	elapsed := stop.Sub(start)                               // Figuring out the time elapsed
	t.Logf("GetPoliciesForUser elapsed time: %v\n", elapsed) // Logging elapsed time

	dobPolicies, err := db.GetPoliciesForUser(contextUser, "dobson")

	//	ASSERT
	if err != nil {
		t.Errorf("GetPoliciesForUser - Should get policies without an error, but got %s", err)
	}

	//	-- Check mal's policies:
	if _, hasPolicy := malPolicies[shipUser.Name]; hasPolicy != true {
		t.Errorf("GetPoliciesForUser - Mal should have '%s', but doesn't.", shipUser.Name)
	}

	if _, hasPolicy := malPolicies[captainAccess.Name]; hasPolicy != true {
		t.Errorf("GetPoliciesForUser - Mal should have '%s', but doesn't.", shipUser.Name)
	}

	if _, hasPolicy := malPolicies[compartmentAccess.Name]; hasPolicy != true {
		t.Errorf("GetPoliciesForUser - Mal should have '%s', but doesn't.", compartmentAccess.Name)
	}

	if _, hasPolicy := malPolicies[healthcareAccess.Name]; hasPolicy != false {
		t.Errorf("GetPoliciesForUser - Mal should NOT have '%s', but does!", healthcareAccess.Name)
	}

	//	-- Check dobson's policies:
	if _, hasPolicy := dobPolicies[healthcareAccess.Name]; hasPolicy != true {
		t.Errorf("GetPoliciesForUser - Dobson should have '%s', but doesn't.", healthcareAccess.Name)
	}

	if _, hasPolicy := dobPolicies[captainAccess.Name]; hasPolicy != false {
		t.Errorf("GetPoliciesForUser - Dobson should NOT have '%s', but does!", captainAccess.Name)
	}

}
