package data_test

import (
	"os"
	"testing"

	"github.com/danesparza/iamserver/data"
)

func TestGroup_AddGroup_ValidGroup_Successful(t *testing.T) {

	//	Arrange
	systemdb, tokendb := getTestFiles()
	db, err := data.NewManager(systemdb, tokendb)
	if err != nil {
		t.Fatalf("NewManager failed: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(systemdb)
		os.RemoveAll(tokendb)
	}()

	contextUser := data.User{Name: "System"}

	//	Act
	response, err := db.AddGroup(contextUser, "UnitTest1", "")

	//	Assert
	if err != nil {
		t.Errorf("AddGroup - Should execute without error, but got: %s", err)
	}

	if response.CreatedBy != contextUser.Name || response.UpdatedBy != contextUser.Name {
		t.Errorf("AddGroup - Should set created and updated by correctly, but got: %s and %s", response.CreatedBy, response.UpdatedBy)
	}

}

func TestGroup_AddGroup_AlreadyExists_ReturnsError(t *testing.T) {

	//	Arrange
	systemdb, tokendb := getTestFiles()
	db, err := data.NewManager(systemdb, tokendb)
	if err != nil {
		t.Fatalf("NewManager failed: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(systemdb)
		os.RemoveAll(tokendb)
	}()

	contextUser := data.User{Name: "System"}

	//	Act
	_, err = db.AddGroup(contextUser, "UnitTest1", "")
	if err != nil {
		t.Errorf("AddGroup - Should execute without error, but got: %s", err)
	}
	_, err = db.AddGroup(contextUser, "UnitTest1", "")

	//	Assert
	if err == nil {
		t.Errorf("AddGroup - Should not add duplicate user without error")
	}

}

func TestGroup_GetGroup_GroupDoesntExist_ReturnsError(t *testing.T) {

	//	Arrange
	systemdb, tokendb := getTestFiles()
	db, err := data.NewManager(systemdb, tokendb)
	if err != nil {
		t.Fatalf("NewManager failed: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(systemdb)
		os.RemoveAll(tokendb)
	}()

	contextUser := data.User{Name: "System"}
	testGroup := "UnitTest1"

	//	Act
	_, err = db.GetGroup(contextUser, testGroup)

	//	Assert
	if err == nil {
		t.Errorf("GetGroup - Should return keynotfound error")
	}

}

func TestGroup_GetGroup_GroupExists_ReturnsGroup(t *testing.T) {

	//	Arrange
	systemdb, tokendb := getTestFiles()
	db, err := data.NewManager(systemdb, tokendb)
	if err != nil {
		t.Fatalf("NewManager failed: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(systemdb)
		os.RemoveAll(tokendb)
	}()

	contextUser := data.User{Name: "System"}
	//	Act
	ret1, err := db.AddGroup(contextUser, "UnitTest1", "")
	if err != nil {
		t.Fatalf("AddGroup - Should execute without error, but got: %s", err)
	}

	_, err = db.AddGroup(contextUser, "UnitTest2", "")
	if err != nil {
		t.Fatalf("AddGroup - Should execute without error, but got: %s", err)
	}

	got1, err := db.GetGroup(contextUser, "UnitTest1")

	//	Assert
	if err != nil {
		t.Errorf("GetGroup - Should get item without error, but got: %s", err)
	}

	if ret1.Name != got1.Name {
		t.Errorf("GetGroup - expected group %s, but got %s instead", "UnitTest1", got1.Name)
	}

}

func TestGroup_AddUsersToGroup_UserDoesntExist_ReturnsError(t *testing.T) {

	//	Arrange
	systemdb, tokendb := getTestFiles()
	db, err := data.NewManager(systemdb, tokendb)
	if err != nil {
		t.Fatalf("NewManager failed: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(systemdb)
		os.RemoveAll(tokendb)
	}()

	contextUser := data.User{Name: "System"}
	adminGroupName := "System admins"

	//	Act

	//	Add some groups
	db.AddGroup(contextUser, adminGroupName, "Unit test administrator group")
	db.AddGroup(contextUser, "Some other group 1", "Unit test group 1")
	db.AddGroup(contextUser, "Some other group 2", "Unit test group 2")
	db.AddGroup(contextUser, "Some other group 3", "Unit test group 3")

	//	Attempt to add users that don't exist yet
	retgrp, err := db.AddUsersToGroup(contextUser, adminGroupName, "nope1@test.com", "nope2@test.com")

	// Sanity check the error
	// t.Logf("AddUsersToGroup error: %s", err)

	if len(retgrp.Users) > 0 {
		t.Errorf("AddUsersToGroup - Should not have added users that don't exist to returned group.  Instead, added %v users", len(retgrp.Users))
	}

	//	Assert
	if err == nil {
		t.Errorf("AddUsersToGroup - Should throw error attempting to add users that don't exist but didn't get an error")
	}

}

func TestGroup_AddUsersToGroup_GroupDoesntExist_ReturnsError(t *testing.T) {

	//	Arrange
	systemdb, tokendb := getTestFiles()
	db, err := data.NewManager(systemdb, tokendb)
	if err != nil {
		t.Fatalf("NewManager failed: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(systemdb)
		os.RemoveAll(tokendb)
	}()

	contextUser := data.User{Name: "System"}
	adminGroupName := "System admins"

	//	Act

	//	NO GROUPS ADDED!

	//	Attempt to add users to group that don't exist yet
	retgrp, err := db.AddUsersToGroup(contextUser, adminGroupName, "nope1@test.com", "nope2@test.com")

	// Sanity check the error
	// t.Logf("AddUsersToGroup error: %s", err)

	if len(retgrp.Users) > 0 {
		t.Errorf("AddUsersToGroup - Should not have added users group that doesn't exist.  Instead, added %v users", len(retgrp.Users))
	}

	//	Assert
	if err == nil {
		t.Errorf("AddUsersToGroup - Should throw error attempting to add users to group that doesn't exist but didn't get an error")
	}

}

func TestGroup_AddUsersToGroup_GroupAndUsersExist_Successful(t *testing.T) {

	//	Arrange
	systemdb, tokendb := getTestFiles()
	db, err := data.NewManager(systemdb, tokendb)
	if err != nil {
		t.Fatalf("NewManager failed: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(systemdb)
		os.RemoveAll(tokendb)
	}()

	contextUser := data.User{Name: "System"}
	adminGroupName := "System admins"

	//	Act

	//	Add some groups
	db.AddGroup(contextUser, adminGroupName, "Unit test administrator group")
	db.AddGroup(contextUser, "Some other group 1", "Unit test group 1")
	db.AddGroup(contextUser, "Some other group 2", "Unit test group 2")
	db.AddGroup(contextUser, "Some other group 3", "Unit test group 3")

	//	Add some users
	db.AddUser(contextUser, data.User{Name: "yep1@test.com"}, "somenewpassword")
	db.AddUser(contextUser, data.User{Name: "yep2@test.com"}, "somenewpassword")
	db.AddUser(contextUser, data.User{Name: "yep3@test.com"}, "somenewpassword")

	//	Attempt to add users exist
	retgrp, err := db.AddUsersToGroup(contextUser, adminGroupName, "yep1@test.com", "yep2@test.com")

	//	Assert
	if err != nil {
		t.Errorf("AddUsersToGroup - should add users without error but got %s", err)
	}

	if len(retgrp.Users) != 2 {
		t.Errorf("AddUsersToGroup - Should have added 2 users.  Instead, added %v users", len(retgrp.Users))
	}

	//	Validate that a lookup on ...
	//  - a 'group' finds the added data:
	retgrp1, err := db.GetGroup(contextUser, adminGroupName)
	if err != nil {
		t.Errorf("GetGroup - should get group without error but got %s", err)
	}

	if retgrp1.Users[1] != "yep2@test.com" {
		t.Logf("GetGroup - expecting 2nd user to be 'yep2@test.com' but got %s instead", retgrp1.Users[1])
	}

	//	- a 'user' finds the added data:
	retusr1, err := db.GetUser(contextUser, "yep1@test.com")
	if err != nil {
		t.Errorf("GetUser - should get user without error but got %s", err)
	}

	if retusr1.Groups[0] != adminGroupName {
		t.Logf("GetUser - expecting 1st group to be '%s' but got %s instead", adminGroupName, retusr1.Groups[0])
	}

}
