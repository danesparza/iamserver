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
		t.Errorf("NewManager failed: %s", err)
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
