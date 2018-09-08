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
		t.Errorf("NewManager failed: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(systemdb)
		os.RemoveAll(tokendb)
	}()

	contextUser := data.User{Name: "System"}
	testGroup := data.Group{Name: "UnitTest1"}

	//	Act
	response, err := db.AddGroup(contextUser, testGroup)

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
		t.Errorf("NewManager failed: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(systemdb)
		os.RemoveAll(tokendb)
	}()

	contextUser := data.User{Name: "System"}
	testGroup := data.Group{Name: "UnitTest1"}

	//	Act
	_, err = db.AddGroup(contextUser, testGroup)
	if err != nil {
		t.Errorf("AddGroup - Should execute without error, but got: %s", err)
	}
	_, err = db.AddGroup(contextUser, testGroup)

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
		t.Errorf("NewManager failed: %s", err)
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
		t.Errorf("NewManager failed: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(systemdb)
		os.RemoveAll(tokendb)
	}()

	contextUser := data.User{Name: "System"}
	testGroup1 := data.Group{Name: "UnitTest1"}
	testGroup2 := data.Group{Name: "UnitTest2"}

	//	Act
	ret1, err := db.AddGroup(contextUser, testGroup1)
	if err != nil {
		t.Fatalf("AddGroup - Should execute without error, but got: %s", err)
	}

	_, err = db.AddGroup(contextUser, testGroup2)
	if err != nil {
		t.Fatalf("AddGroup - Should execute without error, but got: %s", err)
	}

	got1, err := db.GetGroup(contextUser, testGroup1.Name)

	//	Assert
	if err != nil {
		t.Errorf("GetGroup - Should get item without error, but got: %s", err)
	}

	if ret1.Name != got1.Name {
		t.Errorf("GetGroup - expected group %s, but got %s instead", testGroup1.Name, got1.Name)
	}

}
