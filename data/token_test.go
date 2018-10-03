package data_test

import (
	"os"
	"strings"
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

func TestToken_GetUserForToken_ValidParams_Successful(t *testing.T) {

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

	token, err := db.GetNewToken(adminUser, 5*time.Minute)
	if err != nil {
		t.Errorf("GetNewToken - Should execute without error, but got: %s", err)
	}

	//	Act
	userInfo, err := db.GetUserForToken(token.ID)

	//	Assert
	if err != nil {
		t.Errorf("GetUserForToken - Should execute without error, but got: %s", err)
	}

	if userInfo.Name != adminUser.Name {
		t.Errorf("GetUserForToken - Should return correct user, but got: %s", userInfo.Name)
	}

}

func TestToken_GetUserForToken_TokenDoesntExist_ReturnsError(t *testing.T) {

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

	//	Act
	_, err = db.GetUserForToken("BogusID123")

	//	Assert
	if err == nil {
		t.Errorf("GetUserForToken - Should return error for bogus token, but didn't get an error")
	}

}

func TestToken_GetUserForToken_ExpiredToken_ReturnsError(t *testing.T) {

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

	token, err := db.GetNewToken(adminUser, 2*time.Second)
	if err != nil {
		t.Errorf("GetNewToken - Should execute without error, but got: %s", err)
	}

	//	-- Wait for 5 seconds -- TTL should expire and the token should no longer be available:
	time.Sleep(5 * time.Second)

	//	Act
	_, err = db.GetUserForToken(token.ID)

	//	Assert
	if err == nil {
		t.Errorf("GetUserForToken - Should get error for expired token, but didn't get error")
	}

	if !strings.Contains(err.Error(), "Token doesn't exist") {
		t.Errorf("GetUserForToken - Error should be regarding token, but got '%s' instead", err)
	}

}
