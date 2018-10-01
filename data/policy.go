package data

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/danesparza/iamserver/policy"
	"github.com/xtgo/set"

	"github.com/danesparza/badger"
	null "gopkg.in/guregu/null.v3"
	"gopkg.in/guregu/null.v3/zero"
)

// Policy is an AWS style policy document.  They wrap up the following ideas:
// - Resources: The things in a system that users would need permissions to
// - Actions: The interactions users have with those resources
// - Effect: The permissive effect of a policy (allow or deny)
// - Conditions: Additional information to take into account when evaluating a policy
// Policies can be attached to a user or user group.  They can also be grouped in a role
type Policy struct {
	Name      string      `json:"sid"`
	Effect    string      `json:"effect"`
	Resources []string    `json:"resources"`
	Actions   []string    `json:"actions"`
	Roles     []string    `json:"roles"`
	Users     []string    `json:"users"`
	Groups    []string    `json:"groups"`
	Created   time.Time   `json:"created"`
	CreatedBy string      `json:"created_by"`
	Updated   time.Time   `json:"updated"`
	UpdatedBy string      `json:"updated_by"`
	Deleted   zero.Time   `json:"deleted"`
	DeletedBy null.String `json:"deleted_by"`
}

// AddPolicy adds a policy to the system
func (store Manager) AddPolicy(context User, newPolicy Policy) (Policy, error) {
	//	Our return item
	retval := Policy{}

	//	Security check:  Are we authorized to perform this action?
	if !store.IsUserRequestAuthorized(context, sysreqAddPolicy) {
		return retval, fmt.Errorf("User %s is not authorized to perform the action", context.Name)
	}

	//	First -- does the policy exist already?
	err := store.systemdb.View(func(txn *badger.Txn) error {
		_, err := txn.Get(GetKey("Policy", newPolicy.Name))
		return err
	})

	//	If we didn't get an error, we have a problem:
	if err == nil {
		return retval, fmt.Errorf("Policy already exists")
	}

	//	Check Effect
	if (newPolicy.Effect != policy.Allow) && (newPolicy.Effect != policy.Deny) {
		return retval, fmt.Errorf("Policy must have 'allow' or 'deny' effect")
	}

	// 	Check Resources / Actions (they can't be blank or empty)
	if len(newPolicy.Resources) == 0 || len(newPolicy.Actions) == 0 {
		return retval, fmt.Errorf("Policy must have 'resources' and 'actions' associated with it")
	}

	//	Policy must have at least one resource
	if len(newPolicy.Resources) == 0 {
		return retval, fmt.Errorf("Policy must have associated resources (but currently doesn't have any)")
	}

	//	Associated resources have to exist
	for _, currentResource := range newPolicy.Resources {

		//	If the resource name appears to be a regex...
		if strings.ContainsAny(currentResource, "<>") {
			continue // Just go to the next resource
		}

		err := store.systemdb.View(func(txn *badger.Txn) error {
			_, err := txn.Get(GetKey("Resource", currentResource))
			return err
		})

		if err != nil {
			return retval, fmt.Errorf("Resource %s doesn't exist", currentResource)
		}
	}

	//	Policy must have at least one action
	if len(newPolicy.Actions) == 0 {
		return retval, fmt.Errorf("Policy must have associated actions (but currently doesn't have any)")
	}

	//	Make sure when adding a new policy, users / roles / groups are empty:
	newPolicy.Users = []string{}
	newPolicy.Roles = []string{}
	newPolicy.Groups = []string{}

	//	Update the created / updated fields:
	newPolicy.Created = time.Now()
	newPolicy.Updated = time.Now()
	newPolicy.CreatedBy = context.Name
	newPolicy.UpdatedBy = context.Name

	//	Serialize to JSON format
	encoded, err := json.Marshal(newPolicy)
	if err != nil {
		return retval, fmt.Errorf("Problem serializing the data: %s", err)
	}

	//	Save it to the database:
	err = store.systemdb.Update(func(txn *badger.Txn) error {
		err := txn.Set(GetKey("Policy", newPolicy.Name), encoded)
		return err
	})

	//	If there was an error saving the data, report it:
	if err != nil {
		return retval, fmt.Errorf("Problem saving the data: %s", err)
	}

	//	Set our retval:
	retval = newPolicy

	//	Return our data:
	return retval, nil
}

// AttachPolicyToUsers attaches a policy to the given user(s)
func (store Manager) AttachPolicyToUsers(context User, policyName string, users ...string) (Policy, error) {
	//	Our return item
	retval := Policy{}
	affectedUsers := []User{}

	//	Security check:  Are we authorized to perform this action?
	if !store.IsUserRequestAuthorized(context, sysreqAttachPolicyToUsers) {
		return retval, fmt.Errorf("User %s is not authorized to perform the action", context.Name)
	}

	//	First -- validate that the policy exists
	err := store.systemdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(GetKey("Policy", policyName))
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

		return err
	})

	if err != nil {
		return retval, fmt.Errorf("Policy does not exist")
	}

	//	Next:
	//	- Validate that each of the users exist
	//	- Track each user in 'affectedUsers'
	for _, currentuser := range users {
		err := store.systemdb.View(func(txn *badger.Txn) error {
			item, err := txn.Get(GetKey("User", currentuser))

			if err != nil {
				return err
			}

			//	Deserialize the user and add to the list of affected users
			val, err := item.Value()
			if err != nil {
				return err
			}

			if len(val) > 0 {
				currentuserObject := User{}

				//	Unmarshal data into our item
				if err := json.Unmarshal(val, &currentuserObject); err != nil {
					return err
				}

				//	Add the object to our list of affected users:
				affectedUsers = append(affectedUsers, currentuserObject)
			}

			return err
		})

		if err != nil {
			return retval, fmt.Errorf("User %s doesn't exist", currentuser)
		}
	}

	//	Get the policy's new list of users from a merged (and deduped) list of:
	//	- The existing policy users
	// 	- The list of users passed in
	allPolicyUsers := append(retval.Users, users...)
	allUniquePolicyUsers := sort.StringSlice(allPolicyUsers)

	sort.Sort(allUniquePolicyUsers)                 // sort the data first
	n := set.Uniq(allUniquePolicyUsers)             // Uniq returns the size of the set
	allUniquePolicyUsers = allUniquePolicyUsers[:n] // trim the duplicate elements

	//	Then add each of users to the policy
	retval.Users = allUniquePolicyUsers

	//	Serialize the policy to JSON format
	encoded, err := json.Marshal(retval)
	if err != nil {
		return retval, fmt.Errorf("Problem serializing the data: %s", err)
	}

	//	Save it to the database:
	err = store.systemdb.Update(func(txn *badger.Txn) error {
		err := txn.Set(GetKey("Policy", retval.Name), encoded)
		if err != nil {
			return err
		}

		//	Save each affected user to the database (as part of the same transaction):
		for _, affectedUser := range affectedUsers {

			//	and add the group to each user
			currentPolicies := append(affectedUser.Policies, policyName)
			allUniqueCurrentPolicies := sort.StringSlice(currentPolicies)

			sort.Sort(allUniqueCurrentPolicies)                      // sort the data first
			cn := set.Uniq(allUniqueCurrentPolicies)                 // Uniq returns the size of the set
			allUniqueCurrentPolicies = allUniqueCurrentPolicies[:cn] // trim the duplicate elements
			affectedUser.Policies = allUniqueCurrentPolicies

			encoded, err := json.Marshal(affectedUser)
			if err != nil {
				return err
			}

			err = txn.Set(GetKey("User", affectedUser.Name), encoded)
			if err != nil {
				return err
			}
		}

		return nil
	})

	//	If there was an error saving the data, report it:
	if err != nil {
		return retval, fmt.Errorf("Problem completing the policy updates: %s", err)
	}

	//	Return our data:
	return retval, nil

}

// AttachPolicyToGroups attaches a policy to the given group(s)
func (store Manager) AttachPolicyToGroups(context User, policyName string, groups ...string) (Policy, error) {
	//	Our return item
	retval := Policy{}
	affectedGroups := []Group{}

	//	Security check:  Are we authorized to perform this action?
	if !store.IsUserRequestAuthorized(context, sysreqAttachPolicyToGroups) {
		return retval, fmt.Errorf("User %s is not authorized to perform the action", context.Name)
	}

	//	First -- validate that the policy exists
	err := store.systemdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(GetKey("Policy", policyName))
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

		return err
	})

	if err != nil {
		return retval, fmt.Errorf("Policy does not exist")
	}

	//	Next:
	//	- Validate that each of the groups exist
	//	- Track each group in 'affectedGroups'
	for _, currentgroup := range groups {
		err := store.systemdb.View(func(txn *badger.Txn) error {
			item, err := txn.Get(GetKey("Group", currentgroup))

			if err != nil {
				return err
			}

			//	Deserialize the group and add to the list of affected groups
			val, err := item.Value()
			if err != nil {
				return err
			}

			if len(val) > 0 {
				currentgroupObject := Group{}

				//	Unmarshal data into our item
				if err := json.Unmarshal(val, &currentgroupObject); err != nil {
					return err
				}

				//	Add the object to our list of affected groups:
				affectedGroups = append(affectedGroups, currentgroupObject)
			}

			return err
		})

		if err != nil {
			return retval, fmt.Errorf("Group %s doesn't exist", currentgroup)
		}
	}

	//	Get the policy's new list of groups from a merged (and deduped) list of:
	//	- The existing policy groups
	// 	- The list of groups passed in
	allPolicyGroups := append(retval.Groups, groups...)
	allUniquePolicyGroups := sort.StringSlice(allPolicyGroups)

	sort.Sort(allUniquePolicyGroups)                  // sort the data first
	n := set.Uniq(allUniquePolicyGroups)              // Uniq returns the size of the set
	allUniquePolicyGroups = allUniquePolicyGroups[:n] // trim the duplicate elements

	//	Then add each of the groups to the policy
	retval.Groups = allUniquePolicyGroups

	//	Serialize the policy to JSON format
	encoded, err := json.Marshal(retval)
	if err != nil {
		return retval, fmt.Errorf("Problem serializing the data: %s", err)
	}

	//	Save it to the database:
	err = store.systemdb.Update(func(txn *badger.Txn) error {
		err := txn.Set(GetKey("Policy", retval.Name), encoded)
		if err != nil {
			return err
		}

		//	Save each affected group to the database (as part of the same transaction):
		for _, affectedGroup := range affectedGroups {

			//	and add the group to each user
			currentPolicies := append(affectedGroup.Policies, policyName)
			allUniqueCurrentPolicies := sort.StringSlice(currentPolicies)

			sort.Sort(allUniqueCurrentPolicies)                      // sort the data first
			cn := set.Uniq(allUniqueCurrentPolicies)                 // Uniq returns the size of the set
			allUniqueCurrentPolicies = allUniqueCurrentPolicies[:cn] // trim the duplicate elements
			affectedGroup.Policies = allUniqueCurrentPolicies

			encoded, err := json.Marshal(affectedGroup)
			if err != nil {
				return err
			}

			err = txn.Set(GetKey("Group", affectedGroup.Name), encoded)
			if err != nil {
				return err
			}
		}

		return nil
	})

	//	If there was an error saving the data, report it:
	if err != nil {
		return retval, fmt.Errorf("Problem completing the policy updates: %s", err)
	}

	//	Return our data:
	return retval, nil

}

// GetPoliciesForUser gets policies for a user.  Chains include:
// User -> Policies
// User -> Role -> Policies
// User -> Group -> Policies
// User -> Group -> Role -> Policies
func (store Manager) GetPoliciesForUser(context User, userName string) (map[string]Policy, error) {
	//	Our return item
	retval := make(map[string]Policy)
	user := User{}
	policiesInEffect := []string{}
	rolesInEffect := []string{}

	//	First -- validate that the user exists
	err := store.systemdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(GetKey("User", userName))
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}

		if len(val) > 0 {
			//	Unmarshal data into our item
			if err := json.Unmarshal(val, &user); err != nil {
				return err
			}
		}

		return err
	})

	if err != nil {
		return retval, fmt.Errorf("User does not exist")
	}

	//	Add the user policies to policiesInEffect
	for _, currentPolicy := range user.Policies {
		policiesInEffect = append(policiesInEffect, currentPolicy)
	}

	//	Add user roles to rolesInEffect
	for _, currentRole := range user.Roles {
		rolesInEffect = append(rolesInEffect, currentRole)
	}

	//	Find the groups this user is in
	for _, currentGroup := range user.Groups {

		store.systemdb.View(func(txn *badger.Txn) error {
			item, err := txn.Get(GetKey("Group", currentGroup))

			if err != nil {
				return err
			}

			//	Deserialize the group to get the lists of policies and roles
			val, err := item.Value()
			if err != nil {
				return err
			}

			if len(val) > 0 {
				group := Group{}

				//	Unmarshal data into our item
				if err := json.Unmarshal(val, &group); err != nil {
					return err
				}

				//	Add the group policies to policiesInEffect
				for _, currentPolicy := range group.Policies {
					policiesInEffect = append(policiesInEffect, currentPolicy)
				}

				//	Add group roles to rolesInEffect
				for _, currentRole := range group.Roles {
					rolesInEffect = append(rolesInEffect, currentRole)
				}
			}

			return err
		})
	}

	//	Compress and sort the role list
	allUniquerolesInEffect := sort.StringSlice(rolesInEffect)

	sort.Sort(allUniquerolesInEffect)          // sort the data first
	n := set.Uniq(allUniquerolesInEffect)      // Uniq returns the size of the set
	rolesInEffect = allUniquerolesInEffect[:n] // trim the duplicate elements

	//	For each role in rolesInEffect
	for _, currentRole := range rolesInEffect {

		store.systemdb.View(func(txn *badger.Txn) error {
			item, err := txn.Get(GetKey("Role", currentRole))

			if err != nil {
				return err
			}

			//	Deserialize the role to get the list of policies
			val, err := item.Value()
			if err != nil {
				return err
			}

			if len(val) > 0 {
				role := Role{}

				//	Unmarshal data into our item
				if err := json.Unmarshal(val, &role); err != nil {
					return err
				}

				//	Add the role policies to policiesInEffect
				for _, currentPolicy := range role.Policies {
					policiesInEffect = append(policiesInEffect, currentPolicy)
				}
			}

			return err
		})
	}

	//	Compress and sort the policy list
	allUniquepoliciesInEffect := sort.StringSlice(policiesInEffect)

	sort.Sort(allUniquepoliciesInEffect)             // sort the data first
	n = set.Uniq(allUniquepoliciesInEffect)          // Uniq returns the size of the set
	policiesInEffect = allUniquepoliciesInEffect[:n] // trim the duplicate elements

	//	Get the actual policies for each of the policy names
	for _, currentPolicy := range policiesInEffect {

		store.systemdb.View(func(txn *badger.Txn) error {
			item, err := txn.Get(GetKey("Policy", currentPolicy))

			if err != nil {
				return err
			}

			//	Deserialize the role to get the list of policies
			val, err := item.Value()
			if err != nil {
				return err
			}

			if len(val) > 0 {
				policy := Policy{}

				//	Unmarshal data into our item
				if err := json.Unmarshal(val, &policy); err != nil {
					return err
				}

				//	Add the policy to the return value:
				retval[policy.Name] = policy
			}

			return err
		})
	}

	//	Return the list
	return retval, nil
}
