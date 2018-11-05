package data_test

import (
	"fmt"
	"os"
	"path"
	"sort"
	"testing"
	"time"

	"github.com/danesparza/iamserver/data"
	"github.com/danesparza/iamserver/policy"
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
	expectedKey := "User:unitestuser1:name"

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

func TestRoot_Bootstrap_Successful(t *testing.T) {

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

	//	Act
	adminUser, adminSecret, err := db.SystemBootstrap()

	//	Assert
	if err != nil {
		t.Errorf("SystemBootstrap - Should bootstrap without error, but got: %s", err)
	}

	if adminUser.Created.IsZero() || adminUser.Updated.IsZero() {
		t.Errorf("SystemBootstrap failed: Should have set an item with the correct datetime: %+v", adminUser)
	}

	if adminUser.SecretHash == "" {
		t.Errorf("SystemBootstrap failed: Should have set the hashed password correctly: %+v", adminUser)
	}

	if adminSecret == "" {
		t.Errorf("SystemBootstrap failed: Should have returned the admin password correctly but got back blank")
	}

	t.Logf("New Admin user: %+v", adminUser)
	t.Logf("New Admin user secret: %s", adminSecret)

}

func TestRoot_GetOverview_Successful(t *testing.T) {

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

	adminUser, _, err := db.SystemBootstrap()
	if err != nil {
		t.Errorf("SystemBootstrap - Should bootstrap without error, but got: %s", err)
	}

	//	Act
	overview, err := db.GetOverview(adminUser)

	//	Assert
	if err != nil {
		t.Errorf("GetOverview - Should get overview without error, but got: %s", err)
	}

	if overview.GroupCount != 1 {
		t.Errorf("GetOverview - Should get 1 group, but got: %+v", overview)
	}

	if overview.UserCount != 1 {
		t.Errorf("GetOverview - Should get 1 user, but got: %+v", overview)
	}

	if overview.RoleCount != 1 {
		t.Errorf("GetOverview - Should get 1 role, but got: %+v", overview)
	}

	if overview.PolicyCount != 1 {
		t.Errorf("GetOverview - Should get 1 policy, but got: %+v", overview)
	}

	if overview.ResourceCount != 1 {
		t.Errorf("GetOverview - Should get 1 resource, but got: %+v", overview)
	}

}

func TestRoot_GetOverview_HighCapacity_Successful(t *testing.T) {

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

	adminUser, _, err := db.SystemBootstrap()
	if err != nil {
		t.Errorf("SystemBootstrap - Should bootstrap without error, but got: %s", err)
	}

	itemCountToAdd := 10000

	//	-- Add resources:
	for r := 1; r <= itemCountToAdd; r++ {
		db.AddResource(adminUser, fmt.Sprintf("Resource name: %v", r), fmt.Sprintf("Resource desc: %v", r))
	}

	//	-- Add groups:
	for r := 1; r <= itemCountToAdd; r++ {
		db.AddGroup(adminUser, fmt.Sprintf("Group name: %v", r), fmt.Sprintf("Group desc: %v", r))
	}

	//	-- Add roles:
	for r := 1; r <= itemCountToAdd; r++ {
		db.AddRole(adminUser, fmt.Sprintf("Role name: %v", r), fmt.Sprintf("Role desc: %v", r))
	}

	//	-- Add policies:
	for r := 1; r <= itemCountToAdd; r++ {
		db.AddPolicy(adminUser, data.Policy{
			Name:   fmt.Sprintf("Policy name: %v", r),
			Effect: policy.Allow,
			Resources: []string{
				fmt.Sprintf("Resource name: %v", r),
			},
			Actions: []string{
				"Someaction",
			},
		})
	}

	start := time.Now() // Starting the stopwatch

	//	ACT
	overview, err := db.GetOverview(adminUser)

	stop := time.Now()                                                                     // Stopping the stopwatch
	elapsed := stop.Sub(start)                                                             // Figuring out the time elapsed
	t.Logf("GetOverview (%v items each type) elapsed time: %v\n", itemCountToAdd, elapsed) // Logging elapsed time

	//	Assert
	if err != nil {
		t.Errorf("GetOverview - Should get overview without error, but got: %s", err)
	}

	if overview.GroupCount != itemCountToAdd+1 {
		t.Errorf("GetOverview - Should get %v group, but got: %+v", itemCountToAdd+1, overview)
	}

	if overview.UserCount != 1 {
		t.Errorf("GetOverview - Should get 1 user, but got: %+v", overview)
	}

	if overview.RoleCount != itemCountToAdd+1 {
		t.Errorf("GetOverview - Should get %v role, but got: %+v", itemCountToAdd+1, overview)
	}

	if overview.PolicyCount != itemCountToAdd+1 {
		t.Errorf("GetOverview - Should get %v policy, but got: %+v", itemCountToAdd+1, overview)
	}

	if overview.ResourceCount != itemCountToAdd+1 {
		t.Errorf("GetOverview - Should get %v resource, but got: %+v", itemCountToAdd+1, overview)
	}

}

func TestRoot_Search_Successful(t *testing.T) {

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

	adminUser, _, err := db.SystemBootstrap()
	if err != nil {
		t.Errorf("SystemBootstrap - Should bootstrap without error, but got: %s", err)
	}

	//	Act
	results, err := db.Search(adminUser, "admin")

	//	Assert
	if err != nil {
		t.Errorf("Search - Should search without error, but got: %s", err)
	}

	if len(results.Groups) != 1 {
		t.Errorf("Search - Should get 1 group, but got: %+v", results.Groups)
	}

	if len(results.Users) != 1 {
		t.Errorf("Search - Should get 1 user, but got: %+v", results.Users)
	}

	if len(results.Roles) != 1 {
		t.Errorf("Search - Should get 1 role, but got: %+v", results.Roles)
	}

	if len(results.Policies) != 1 {
		t.Errorf("Search - Should get 1 policy, but got: %+v", results.Policies)
	}

	if len(results.Resources) != 0 {
		t.Errorf("Search - Should get 0 resources, but got: %+v", results.Resources)
	}

}

func TestRoot_Search_HighCapacity_Successful(t *testing.T) {

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

	adminUser, _, err := db.SystemBootstrap()
	if err != nil {
		t.Errorf("SystemBootstrap - Should bootstrap without error, but got: %s", err)
	}

	itemCountToAdd := 10000
	itemCountExpected := 3439

	//	-- Add resources:
	for r := 1; r <= itemCountToAdd; r++ {
		db.AddResource(adminUser, fmt.Sprintf("UnitTestResource%v", r), fmt.Sprintf("Resource desc: %v", r))
	}

	//	-- Add groups:
	for r := 1; r <= itemCountToAdd; r++ {
		db.AddGroup(adminUser, fmt.Sprintf("UnitTestGroup%v", r), fmt.Sprintf("Group desc: %v", r))
	}

	//	-- Add roles:
	for r := 1; r <= itemCountToAdd; r++ {
		db.AddRole(adminUser, fmt.Sprintf("UnitTestRole%v", r), fmt.Sprintf("Role desc: %v", r))
	}

	//	-- Add policies:
	for r := 1; r <= itemCountToAdd; r++ {
		db.AddPolicy(adminUser, data.Policy{
			Name:   fmt.Sprintf("UnitTestPolicy%v", r),
			Effect: policy.Allow,
			Resources: []string{
				fmt.Sprintf("UnitTestResource%v", r),
			},
			Actions: []string{
				"Someaction",
			},
		})
	}

	start := time.Now() // Starting the stopwatch

	//	ACT
	results, err := db.Search(adminUser, "unittest.*2")

	stop := time.Now()                                                                // Stopping the stopwatch
	elapsed := stop.Sub(start)                                                        // Figuring out the time elapsed
	t.Logf("Search (%v items each type) elapsed time: %v\n", itemCountToAdd, elapsed) // Logging elapsed time

	//	Assert
	if err != nil {
		t.Errorf("Search - Should get search results without error, but got: %s", err)
	}

	if len(results.Groups) != itemCountExpected {
		t.Errorf("Search - Should get %v groups, but got: %+v", itemCountExpected, len(results.Groups))
	}

	if len(results.Users) != 0 {
		t.Errorf("Search - Should get 1 user, but got: %+v", len(results.Users))
	}

	if len(results.Roles) != itemCountExpected {
		t.Errorf("Search - Should get %v roles, but got: %+v", itemCountExpected, len(results.Roles))
	}

	if len(results.Policies) != itemCountExpected {
		t.Errorf("Search - Should get %v policies, but got: %+v", itemCountExpected, len(results.Policies))
	}

	if len(results.Resources) != itemCountExpected {
		t.Errorf("Search - Should get %v resources, but got: %+v", itemCountExpected, len(results.Resources))
	}

}
