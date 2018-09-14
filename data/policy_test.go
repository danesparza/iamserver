package data_test

import (
	"os"
	"testing"

	"github.com/danesparza/iamserver/data"
	"github.com/danesparza/iamserver/policy"
)

func TestPolicy_AddPolicy_ValidPolicy_Successful(t *testing.T) {

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
	testPolicy := data.Policy{
		Name:   "UnitTest1",
		Effect: policy.Allow,
		Resources: []string{
			"Someresource",
		},
		Actions: []string{
			"Someaction",
		},
	}

	//	Act
	newPolicy, err := db.AddPolicy(contextUser, testPolicy)

	//	Assert
	if err != nil {
		t.Errorf("AddPolicy - Should add policy without error, but got: %s", err)
	}

	if newPolicy.Created.IsZero() || newPolicy.Updated.IsZero() {
		t.Errorf("AddPolicy failed: Should have set an item with the correct datetime: %+v", newPolicy)
	}

	if newPolicy.CreatedBy != contextUser.Name {
		t.Errorf("AddPolicy failed: Should have set an item with the correct 'created by' user: %+v", newPolicy)
	}

	if newPolicy.UpdatedBy != contextUser.Name {
		t.Errorf("AddPolicy failed: Should have set an item with the correct 'updated by' user: %+v", newPolicy)
	}
}

func TestPolicy_AddPolicy_AlreadyExists_ReturnsError(t *testing.T) {

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
	testPolicy := data.Policy{
		Name:   "UnitTest1",
		Effect: policy.Allow,
		Resources: []string{
			"Someresource",
		},
		Actions: []string{
			"Someaction",
		},
	}

	//	Act
	_, err = db.AddPolicy(contextUser, testPolicy)
	if err != nil {
		t.Errorf("AddPolicy - Should add policy without error, but got: %s", err)
	}
	_, err = db.AddPolicy(contextUser, testPolicy)

	//	Assert
	if err == nil {
		t.Errorf("AddPolicy - Should not add duplicate policy without error")
	}

}

func TestPolicy_AddPolicy_InvalidEffect_ReturnsError(t *testing.T) {

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
	testPolicy := data.Policy{
		Name:   "UnitTest1",
		Effect: "someweirdeffect",
		Resources: []string{
			"Someresource",
		},
		Actions: []string{
			"Someaction",
		},
	}

	//	Act
	_, err = db.AddPolicy(contextUser, testPolicy)

	//	Assert
	if err == nil {
		t.Errorf("AddPolicy - Should not add policy with invalid effect")
	}

}
