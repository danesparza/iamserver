package data_test

import (
	"os"
	"testing"

	"github.com/danesparza/iamserver/data"
)

func TestRole_AddRole_ValidRole_Successful(t *testing.T) {

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
	response, err := db.AddRole(contextUser, "UnitTest1", "")

	//	Assert
	if err != nil {
		t.Errorf("AddRole - Should execute without error, but got: %s", err)
	}

	if response.CreatedBy != contextUser.Name || response.UpdatedBy != contextUser.Name {
		t.Errorf("AddRole - Should set created and updated by correctly, but got: %s and %s", response.CreatedBy, response.UpdatedBy)
	}

}

func TestRole_AddRole_AlreadyExists_ReturnsError(t *testing.T) {

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
	_, err = db.AddRole(contextUser, "UnitTest1", "")
	if err != nil {
		t.Errorf("AddRole - Should execute without error, but got: %s", err)
	}
	_, err = db.AddRole(contextUser, "UnitTest1", "")

	//	Assert
	if err == nil {
		t.Errorf("AddRole - Should not add duplicate user without error")
	}

}

func TestRole_GetRole_RoleDoesntExist_ReturnsError(t *testing.T) {

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
	testRole := "UnitTest1"

	//	Act
	_, err = db.GetRole(contextUser, testRole)

	//	Assert
	if err == nil {
		t.Errorf("GetRole - Should return keynotfound error")
	}

}

func TestRole_GetRole_RoleExists_ReturnsRole(t *testing.T) {

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
	ret1, err := db.AddRole(contextUser, "UnitTest1", "")
	if err != nil {
		t.Fatalf("AddRole - Should execute without error, but got: %s", err)
	}

	_, err = db.AddRole(contextUser, "UnitTest2", "")
	if err != nil {
		t.Fatalf("AddRole - Should execute without error, but got: %s", err)
	}

	got1, err := db.GetRole(contextUser, "UnitTest1")

	//	Assert
	if err != nil {
		t.Errorf("GetRole - Should get item without error, but got: %s", err)
	}

	if ret1.Name != got1.Name {
		t.Errorf("GetRole - expected group %s, but got %s instead", "UnitTest1", got1.Name)
	}

}

func TestRole_AddPoliciesToRole_PolicyDoesntExist_ReturnsError(t *testing.T) {

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
	adminRoleName := "Administrator role"

	//	Act

	//	Add some roles
	db.AddRole(contextUser, adminRoleName, "Unit test administrator role")
	db.AddRole(contextUser, "Some other role 1", "Unit test role 1")
	db.AddRole(contextUser, "Some other role 2", "Unit test role 2")
	db.AddRole(contextUser, "Some other role 3", "Unit test role 3")

	//	Attempt to add policies that don't exist yet
	retrole, err := db.AddPoliciesToRole(contextUser, adminRoleName, "policy 1", "policy 2")

	// Sanity check the error
	// t.Logf("AddPoliciesToRole error: %s", err)

	if len(retrole.Policies) > 0 {
		t.Errorf("AddPoliciesToRole - Should not have added policies that don't exist to returned role.  Instead, added %v policies", len(retrole.Policies))
	}

	//	Assert
	if err == nil {
		t.Errorf("AddPoliciesToRole - Should throw error attempting to add policies that don't exist but didn't get an error")
	}

}

func TestRole_AddPoliciesToRole_RoleDoesntExist_ReturnsError(t *testing.T) {

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
	adminRoleName := "Administrator role"

	//	Act

	//	NO ROLES ADDED!

	//	Attempt to add users to group that don't exist yet
	retrole, err := db.AddPoliciesToRole(contextUser, adminRoleName, "policy 1", "policy 2")

	// Sanity check the error
	// t.Logf("AddUsersToGroup error: %s", err)

	if len(retrole.Policies) > 0 {
		t.Errorf("AddPoliciesToRole - Should not have added policies to role that doesn't exist.  Instead, added %v policies", len(retrole.Policies))
	}

	//	Assert
	if err == nil {
		t.Errorf("AddPoliciesToRole - Should throw error attempting to add policies to role that doesn't exist but didn't get an error")
	}

}

func TestRole_AttachRoleToUsers_PolicyDoesntExist_ReturnsError(t *testing.T) {

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

	//	Attempt to attach roles that don't exist yet
	retrole, err := db.AttachRoleToUsers(contextUser, "Bad role 1", "Unittestuser1", "Unittestuser2", "Unittestuser3")

	// Sanity check the error
	t.Logf("AttachRoleToUsers error: %s", err)

	if len(retrole.Users) > 0 {
		t.Errorf("AttachRoleToUsers - Should not have attached roles that don't exist.")
	}

	//	Assert
	if err == nil {
		t.Errorf("AttachRoleToUsers - Should throw error attempting to attach roles that don't exist but didn't get an error")
	}
}

func TestRole_AttachRoleToGroups_PolicyDoesntExist_ReturnsError(t *testing.T) {

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

	//	Attempt to attach roles that don't exist yet
	retrole, err := db.AttachRoleToGroups(contextUser, "Bad role 1", "Unittestgroup1", "Unittestgroup2", "Unittestgroup3")

	// Sanity check the error
	t.Logf("AttachRoleToGroups error: %s", err)

	if len(retrole.Groups) > 0 {
		t.Errorf("AttachRoleToGroups - Should not have attached roles that don't exist.")
	}

	//	Assert
	if err == nil {
		t.Errorf("AttachRoleToGroups - Should throw error attempting to attach roles that don't exist but didn't get an error")
	}

}

func TestRole_AttachRoleToUsers_UserDoesntExist_ReturnsError(t *testing.T) {

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

	//	Add a role
	newRole, _ := db.AddRole(contextUser, "UnitTest1", "")

	//	Act

	//	Attempt to attach role to users that don't exist yet
	retrole, err := db.AttachRoleToUsers(contextUser, newRole.Name, "Unittestuser1", "Unittestuser2", "Unittestuser3")

	// Sanity check the error
	// t.Logf("AttachRoleToUsers error: %s", err)

	if len(retrole.Users) > 0 {
		t.Errorf("AttachRoleToUsers - Should not have attached role to users that don't exist.")
	}

	//	Assert
	if err == nil {
		t.Errorf("AttachRoleToUsers - Should throw error attempting to attach role to users that don't exist but didn't get an error")
	}
}

func TestRole_AttachRoleToGroups_GroupDoesntExist_ReturnsError(t *testing.T) {

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

	//	Add a role
	newRole, _ := db.AddRole(contextUser, "UnitTest1", "")

	//	Act

	//	Attempt to attach role to groups that don't exist yet
	retrole, err := db.AttachRoleToGroups(contextUser, newRole.Name, "Unittestgroup1", "Unittestgroup2", "Unittestgroup3")

	// Sanity check the error
	// t.Logf("AttachRoleToGroups error: %s", err)

	if len(retrole.Groups) > 0 {
		t.Errorf("AttachRoleToGroups - Should not have attached role to groups that don't exist.")
	}

	//	Assert
	if err == nil {
		t.Errorf("AttachRoleToGroups - Should throw error attempting to attach role to groups that don't exist but didn't get an error")
	}

}
