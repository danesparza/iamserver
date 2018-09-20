package data_test

import (
	"os"
	"testing"

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

	if len(retpolicy.Users) > 0 {
		t.Errorf("AttachPolicyToGroups - Should not have attached policies that don't exist.")
	}

	//	Assert
	if err == nil {
		t.Errorf("AttachPolicyToGroups - Should throw error attempting to attach policies that don't exist but didn't get an error")
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
