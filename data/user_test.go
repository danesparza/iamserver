package data_test

import (
	"os"
	"testing"

	"github.com/danesparza/iamserver/data"
)

func TestUser_AddUser_ValidUser_Successful(t *testing.T) {

	//	Arrange
	systemdb, tokendb := getTestFiles()
	db := data.Manager{SystemDBpath: systemdb, TokenDBpath: tokendb}
	defer func() {
		os.RemoveAll(systemdb)
	}()

	contextUser := data.User{Name: "System"}
	testUser := data.User{Name: "UnitTest1"}
	testPassword := "testpass"

	//	Act
	_, err := db.AddUser(contextUser, testUser, testPassword)

	//	Assert
	if err != nil {
		t.Errorf("AddUser - Should add user without error, but got: %s", err)
	}

}

func TestUser_GetUser_UserDoesntExist_ReturnsError(t *testing.T) {

	//	Arrange
	systemdb, tokendb := getTestFiles()
	db := data.Manager{SystemDBpath: systemdb, TokenDBpath: tokendb}
	defer func() {
		os.RemoveAll(systemdb)
	}()

	contextUser := data.User{Name: "System"}
	testUser := "UnitTest1"

	//	Act
	_, err := db.GetUser(contextUser, testUser)

	//	Assert
	if err == nil {
		t.Errorf("GetUser - Should return keynotfound error")
	}

}
