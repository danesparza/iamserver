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
		t.Fatalf("NewManager failed: %s", err)
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
	newUser, err := db.AddUser(contextUser, testUser, testPassword)

	//	Assert
	if err != nil {
		t.Errorf("AddUser - Should add user without error, but got: %s", err)
	}

	if newUser.Created.IsZero() || newUser.Updated.IsZero() {
		t.Errorf("AddUser failed: Should have set an item with the correct datetime: %+v", newUser)
	}

	if newUser.CreatedBy != contextUser.Name {
		t.Errorf("AddUser failed: Should have set an item with the correct 'created by' user: %+v", newUser)
	}

	if newUser.UpdatedBy != contextUser.Name {
		t.Errorf("AddUser failed: Should have set an item with the correct 'updated by' user: %+v", newUser)
	}

	if newUser.SecretHash == "" || newUser.SecretHash == testPassword {
		t.Errorf("AddUser failed: Should have set the hashed password correctly: %+v", newUser)
	}

}

func TestUser_AddUser_AlreadyExists_ReturnsError(t *testing.T) {

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

func TestUser_AddUser_XSSAttempt_SanitizesInput(t *testing.T) {

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

	var userName = string([]byte(`TestUser<img src=x onerror="alert('Pop-up window via stored XSS');"`))
	testUser := data.User{Name: userName}
	testPassword := "testpass"

	//	Act
	newUser, err := db.AddUser(contextUser, testUser, testPassword)

	//	Assert
	if err != nil {
		t.Errorf("AddUser - Should add user without error, but got: %s", err)
	}

	if newUser.Name != "TestUser" {
		t.Errorf("AddUser failed: Should have sanitized the username but got: %+v", newUser)
	}

	if newUser.Created.IsZero() || newUser.Updated.IsZero() {
		t.Errorf("AddUser failed: Should have set an item with the correct datetime: %+v", newUser)
	}

	if newUser.CreatedBy != contextUser.Name {
		t.Errorf("AddUser failed: Should have set an item with the correct 'created by' user: %+v", newUser)
	}

	if newUser.UpdatedBy != contextUser.Name {
		t.Errorf("AddUser failed: Should have set an item with the correct 'updated by' user: %+v", newUser)
	}

	if newUser.SecretHash == "" || newUser.SecretHash == testPassword {
		t.Errorf("AddUser failed: Should have set the hashed password correctly: %+v", newUser)
	}

}

func TestUser_GetUser_UserDoesntExist_ReturnsError(t *testing.T) {

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
		t.Fatalf("NewManager failed: %s", err)
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
		t.Fatalf("NewManager failed: %s", err)
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

func TestUser_GetUserWithCredentials_ValidParams_ReturnsUser(t *testing.T) {

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
	testUser1 := data.User{Name: "UnitTest1"}
	testPassword := "testpass"

	_, err = db.AddUser(contextUser, testUser1, testPassword)
	if err != nil {
		t.Fatalf("AddUser - Should add user without error, but got: %s", err)
	}

	//	Act
	gotuser1, err := db.GetUserWithCredentials(testUser1.Name, testPassword)

	//	Assert
	if err != nil {
		t.Errorf("GetUserWithCredentials - Should get user without error, but got: %s", err)
	}

	if testUser1.Name != gotuser1.Name {
		t.Errorf("GetUserWithCredentials - expected user %s, but got %s instead", testUser1.Name, gotuser1.Name)
	}

}

func TestUser_GetUserWithCredentials_InvalidParams_ReturnsError(t *testing.T) {

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
	testUser1 := data.User{Name: "UnitTest1"}
	testPassword := "testpass"

	_, err = db.AddUser(contextUser, testUser1, testPassword)
	if err != nil {
		t.Fatalf("AddUser - Should add user without error, but got: %s", err)
	}

	//	Act
	gotuser1, err := db.GetUserWithCredentials(testUser1.Name, "INCORRECT_PASSWORD")

	//	Assert
	if err == nil {
		t.Errorf("GetUserWithCredentials - Should have gotten error for incorrect password, but didn't")
	}

	if testUser1.Name == gotuser1.Name {
		t.Errorf("GetUserWithCredentials - should NOT have gotten user information back, but did")
	}

}
