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
