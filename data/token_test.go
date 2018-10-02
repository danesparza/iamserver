package data_test

import (
	"os"
	"testing"
	"time"

	"github.com/danesparza/iamserver/data"
)

func TestToken_GetNewToken_UserDoesntExist_ReturnsError(t *testing.T) {

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

	testUser := data.User{Name: "TestUser"}

	//	Act
	_, err = db.GetNewToken(testUser, 5*time.Minute)

	//	Assert
	if err == nil {
		t.Errorf("GetNewToken - Should throw an error because the user doens't exist, but didn't get one")
	}

}

func TestToken_GetNewToken_ValidParams_Successful(t *testing.T) {

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

	adminUser, _, err := db.SystemBootstrap()
	if err != nil {
		t.Errorf("SystemBootstrap - Should execute without error, but got: %s", err)
	}

	//	Act
	token, err := db.GetNewToken(adminUser, 5*time.Minute)

	//	Assert
	if err != nil {
		t.Errorf("GetNewToken - Should execute without error, but got: %s", err)
	}

	if err == nil {
		t.Logf("Got new token: %+v", token)
	}

}
