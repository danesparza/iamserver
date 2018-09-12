package data

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/danesparza/badger"
	null "gopkg.in/guregu/null.v3"
	"gopkg.in/guregu/null.v3/zero"
)

// Resource represents a thing that can be acted on.  This is really only used for lookups when
// editing a policy.  Because a policy can have wildcards, this type isn't used for policy validation.
type Resource struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Created     time.Time        `json:"created"`
	CreatedBy   string           `json:"created_by"`
	Updated     time.Time        `json:"updated"`
	UpdatedBy   string           `json:"updated_by"`
	Deleted     zero.Time        `json:"deleted"`
	DeletedBy   null.String      `json:"deleted_by"`
	Actions     []ResourceAction `json:"actions"`
}

// AddResource adds a resource to the system
func (store Manager) AddResource(context User, name, description string) (Resource, error) {
	//	Our return item
	retval := Resource{}
	newResource := Resource{Name: name, Description: description}

	//	First -- does the resource exist already?
	err := store.systemdb.View(func(txn *badger.Txn) error {
		_, err := txn.Get(GetKey("Resource", newResource.Name))
		return err
	})

	//	If we didn't get an error, we have a problem:
	if err == nil {
		return retval, fmt.Errorf("Resource already exists")
	}

	//	Update the created / updated fields:
	newResource.Created = time.Now()
	newResource.Updated = time.Now()
	newResource.CreatedBy = context.Name
	newResource.UpdatedBy = context.Name

	//	Serialize to JSON format
	encoded, err := json.Marshal(newResource)
	if err != nil {
		return retval, fmt.Errorf("Problem serializing the data: %s", err)
	}

	//	Save it to the database:
	err = store.systemdb.Update(func(txn *badger.Txn) error {
		err := txn.Set(GetKey("Resource", newResource.Name), encoded)
		return err
	})

	//	If there was an error saving the data, report it:
	if err != nil {
		return retval, fmt.Errorf("Problem saving the data: %s", err)
	}

	//	Set our retval:
	retval = newResource

	//	Return our data:
	return retval, nil
}

// GetResource gets a resource from the system
func (store Manager) GetResource(context User, resourceName string) (Resource, error) {
	//	Our return item
	retval := Resource{}

	err := store.systemdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(GetKey("Resource", resourceName))
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}

		if len(val) > 0 {
			//	Unmarshal data into our item
			if err := json.Unmarshal(val, &retval); err != nil {
				return err
			}
		}

		return nil
	})

	//	If there was an error, report it:
	if err != nil {
		return retval, fmt.Errorf("Problem getting the data: %s", err)
	}

	//	Return our data:
	return retval, nil
}

// GetAllResources gets all resources in the system
func (store Manager) GetAllResources(context User) ([]Resource, error) {
	//	Our return item
	retval := []Resource{}

	err := store.systemdb.View(func(txn *badger.Txn) error {

		//	Get an iterator
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		//	Set our prefix
		prefix := GetKey("Resource")

		//	Iterate over our values:
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {

			//	Get the item
			item := it.Item()

			//	Get the item key
			// k := item.Key()

			//	Get the item value
			val, err := item.Value()
			if err != nil {
				return err
			}

			if len(val) > 0 {
				//	Create our item:
				item := Resource{}

				//	Unmarshal data into our item
				if err := json.Unmarshal(val, &item); err != nil {
					return err
				}

				//	Add to the array of returned users:
				retval = append(retval, item)
			}
		}
		return nil
	})

	//	If there was an error, report it:
	if err != nil {
		return retval, fmt.Errorf("Problem getting the list of items: %s", err)
	}

	//	Return our data:
	return retval, nil
}
