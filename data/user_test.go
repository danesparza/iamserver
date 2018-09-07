package data_test

import (
	"os"
	"testing"

	"github.com/danesparza/iamserver/data"
)

func TestUser_AddUser_ValidUser_Successful(t *testing.T) {

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
	testUser := data.User{Name: "UnitTest1"}
	testPassword := "testpass"

	//	Act
	_, err = db.AddUser(contextUser, testUser, testPassword)

	//	Assert
	if err != nil {
		t.Errorf("AddUser - Should add user without error, but got: %s", err)
	}

}

func TestUser_AddUser_AlreadyExists_ReturnsError(t *testing.T) {

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
	testUser := data.User{Name: "UnitTest1"}
	testPassword := "testpass"

	//	Act
	_, err = db.AddUser(contextUser, testUser, testPassword)
	if err != nil {
		t.Errorf("AddUser - Should add user without error, but got: %s", err)
	}
	_, err = db.AddUser(contextUser, testUser, testPassword)

	//	Assert
	if err == nil {
		t.Errorf("AddUser - Should not add duplicate user without error")
	}

}

func TestUser_GetUser_UserDoesntExist_ReturnsError(t *testing.T) {

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
	testUser := "UnitTest1"

	//	Act
	_, err = db.GetUser(contextUser, testUser)

	//	Assert
	if err == nil {
		t.Errorf("GetUser - Should return keynotfound error")
	}

}

func TestUser_GetUser_UserExists_ReturnsUser(t *testing.T) {

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
	testUser1 := data.User{Name: "UnitTest1"}
	testUser2 := data.User{Name: "UnitTest2"}
	testPassword := "testpass"

	//	Act
	retuser1, err := db.AddUser(contextUser, testUser1, testPassword)
	if err != nil {
		t.Fatalf("AddUser - Should add user without error, but got: %s", err)
	}

	_, err = db.AddUser(contextUser, testUser2, testPassword)
	if err != nil {
		t.Fatalf("AddUser - Should add user without error, but got: %s", err)
	}

	gotuser1, err := db.GetUser(contextUser, testUser1.Name)

	//	Assert
	if err != nil {
		t.Errorf("GetUser - Should get user without error, but got: %s", err)
	}

	if retuser1.Name != gotuser1.Name {
		t.Errorf("GetUser - expected user %s, but got %s instead", retuser1.Name, gotuser1.Name)
	}

}

func TestUser_GetAllUsers_UserExists_ReturnsAllUsers(t *testing.T) {

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
	testUser1 := data.User{Name: "UnitTest1"}
	testUser2 := data.User{Name: "UnitTest2"}
	testPassword := "testpass"

	//	Act
	retuser1, err := db.AddUser(contextUser, testUser1, testPassword)
	if err != nil {
		t.Fatalf("AddUser - Should add user without error, but got: %s", err)
	}

	_, err = db.AddUser(contextUser, testUser2, testPassword)
	if err != nil {
		t.Fatalf("AddUser - Should add user without error, but got: %s", err)
	}

	allusers, err := db.GetAllUsers(contextUser)

	//	Assert
	if err != nil {
		t.Errorf("GetAllUsers - Should get all users without error, but got: %s", err)
	}

	if len(allusers) != 2 {
		t.Errorf("GetAllUsers - expected 2 users, but got %v instead", len(allusers))
	}

	if allusers[0].Name != retuser1.Name {
		t.Errorf("GetAllUsers - expected first user to be %s, but got %s instead", retuser1.Name, allusers[0].Name)
	}

}
