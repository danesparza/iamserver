package data_test

import (
	"os"
	"testing"

	"github.com/danesparza/iamserver/data"
)

func TestResource_AddResource_ValidResource_Successful(t *testing.T) {

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
	testResource := data.Resource{Name: "ResourceTest1", Description: "Test description"}

	//	Act
	newResource, err := db.AddResource(contextUser, testResource.Name, testResource.Description)

	//	Assert
	if err != nil {
		t.Errorf("AddResource - Should add resource without error, but got: %s", err)
	}

	if newResource.Created.IsZero() || newResource.Updated.IsZero() {
		t.Errorf("AddResource failed: Should have set an item with the correct datetime: %+v", newResource)
	}

	if newResource.CreatedBy != contextUser.Name {
		t.Errorf("AddResource failed: Should have set an item with the correct 'created by' user: %+v", newResource)
	}

	if newResource.UpdatedBy != contextUser.Name {
		t.Errorf("AddResource failed: Should have set an item with the correct 'updated by' user: %+v", newResource)
	}

}

func TestResource_AddResource_AlreadyExists_ReturnsError(t *testing.T) {

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
	testResource := data.Resource{Name: "ResourceTest1", Description: "Test description"}

	//	Act
	_, err = db.AddResource(contextUser, testResource.Name, testResource.Description)
	if err != nil {
		t.Errorf("AddResource - Should add resource without error, but got: %s", err)
	}

	_, err = db.AddResource(contextUser, testResource.Name, testResource.Description)

	//	Assert
	if err == nil {
		t.Errorf("AddResource - Should not add duplicate without error")
	}

}

func TestResource_AddActionsToResource_ResourceDoesntExist_ReturnsError(t *testing.T) {

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
	resourceName := "SomeNewApplication"

	//	Act

	//	Attempt to add actions to a resource that doesn't exist:
	_, err = db.AddActionToResource(contextUser, resourceName, "GetForm", "ListForms")

	// Sanity check the error
	// t.Logf("AddActionToResource error: %s", err)

	//	Assert
	if err == nil {
		t.Errorf("AddActionToResource - Should throw error attempting to add actions to a resource that doesn't exist but didn't get an error")
	}

}
