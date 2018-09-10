package data_test

import (
	"os"
	"path"
	"sort"
	"testing"

	"github.com/danesparza/iamserver/data"
	"github.com/xtgo/set"
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
	actualKey := data.GetKey("User", userId, "name")

	//	Assert
	if expectedKey != string(actualKey) {
		t.Errorf("GetKey failed:  Expected %s but got %s instead", expectedKey, actualKey)
	}
}

func TestUniq(t *testing.T) {

	//	Our regular slice o' emails
	emails := []string{"esparza.dan@gmail.com", "danesparza@cagedtornado.com", "cmesparza@gmail.com", "esparza.dan@gmail.com"}

	//	Convert them into a sortable slice:
	data := sort.StringSlice(emails)

	sort.Sort(data)     // sort the data first
	n := set.Uniq(data) // Uniq returns the size of the set
	data = data[:n]     // trim the duplicate elements

	//	 Well looky here ... we have a unique (sorted) set of emails
	// t.Logf("%+v", data)

	if len(data) > 3 {
		t.Errorf("Expecting only 3 unique elements, but found: %v", len(data))
	}

}
