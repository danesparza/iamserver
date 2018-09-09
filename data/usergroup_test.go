package data_test

import (
	"os"
	"sort"
	"testing"

	"github.com/danesparza/iamserver/data"
	"github.com/xtgo/set"
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

	//	Act
	response, err := db.AddGroup(contextUser, "UnitTest1", "")

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

	//	Act
	_, err = db.AddGroup(contextUser, "UnitTest1", "")
	if err != nil {
		t.Errorf("AddGroup - Should execute without error, but got: %s", err)
	}
	_, err = db.AddGroup(contextUser, "UnitTest1", "")

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
	//	Act
	ret1, err := db.AddGroup(contextUser, "UnitTest1", "")
	if err != nil {
		t.Fatalf("AddGroup - Should execute without error, but got: %s", err)
	}

	_, err = db.AddGroup(contextUser, "UnitTest2", "")
	if err != nil {
		t.Fatalf("AddGroup - Should execute without error, but got: %s", err)
	}

	got1, err := db.GetGroup(contextUser, "UnitTest1")

	//	Assert
	if err != nil {
		t.Errorf("GetGroup - Should get item without error, but got: %s", err)
	}

	if ret1.Name != got1.Name {
		t.Errorf("GetGroup - expected group %s, but got %s instead", "UnitTest1", got1.Name)
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
	t.Logf("%+v", data)

}
