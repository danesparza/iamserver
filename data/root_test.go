package data_test

import (
	"os"
	"path"
	"testing"

	"github.com/danesparza/iamserver/data"
)

//	Gets the database path for this environment:
func getTestFiles() (string, string) {
	systemdb := ""
	tokendb := ""

	testRoot := os.Getenv("IAM_TEST_ROOT")

	if testRoot != "" {
		systemdb = path.Join(testRoot, "system")
		tokendb = path.Join(testRoot, "token")
	}

	return systemdb, tokendb
}

func TestRoot_GetTestDBPaths_Successful(t *testing.T) {

	systemdb, tokendb := getTestFiles()

	if systemdb == "" || tokendb == "" {
		t.Fatal("The required IAM_TEST_ROOT environment variable is not set to the test database root path")
	}

	t.Logf("System db path: %s", systemdb)
	t.Logf("Token db path: %s", tokendb)
}

func TestRoot_Databases_ShouldNotExistYet(t *testing.T) {
	//	Arrange
	systemdb, tokendb := getTestFiles()

	//	Act

	//	Assert
	if _, err := os.Stat(systemdb); err == nil {
		t.Errorf("System database check failed: System db directory %s already exists, and shouldn't", systemdb)
	}

	if _, err := os.Stat(tokendb); err == nil {
		t.Errorf("Token database check failed: Token db directory %s already exists, and shouldn't", tokendb)
	}
}

func TestRoot_GetKey_ReturnsCorrectKey(t *testing.T) {
	//	Arrange
	userId := "unitestuser1"
	expectedKey := "User_unitestuser1_name"

	//	Act
	db := data.Manager{}
	actualKey := db.GetKey("User", userId, "name")

	//	Assert
	if expectedKey != actualKey {
		t.Errorf("GetKey failed:  Expected %s but got %s instead", expectedKey, actualKey)
	}
}
