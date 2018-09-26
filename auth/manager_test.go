package auth_test

import (
	"testing"

	"github.com/danesparza/iamserver/auth"
	"github.com/danesparza/iamserver/data"
	"github.com/danesparza/iamserver/policy"
)

func TestManager_DoPoliciesAllow_ValidRequest_Successful(t *testing.T) {

	//	Arrange
	mgr := &auth.Manager{}
	pols := []data.Policy{
		{
			Name:   "Regular user ship access",
			Effect: policy.Allow,
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"Find",
				"Open",
				"Embark",
				"Disembark",
			},
		},
		{
			Name:   "Captain privledges",
			Effect: policy.Allow,
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"Fly",
				"Navigate",
				"Curse",
			},
		},
		{
			Name:   "Secret compartment access",
			Effect: policy.Allow,
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"AccessCompartments",
			},
		},
		{
			Name:   "Healthcare access",
			Effect: policy.Allow,
			Resources: []string{
				"Healthcare",
			},
			Actions: []string{
				"PresentHMOcard",
				"WaitToSeeDoc",
				"GetMedicalAdvice",
			},
		},
	}

	req := &data.Request{
		Action:   "Embark",
		Resource: "Serenity",
		User:     "malreynolds",
	}

	//	Act
	err := mgr.DoPoliciesAllow(req, pols)

	//	Assert
	if err != nil {
		t.Errorf("DoPoliciesAllow - should allow request, but got error: %v", err)
	}

}

func TestManager_DoPoliciesAllow_InvalidRequest_ReturnsError(t *testing.T) {

	//	Arrange
	mgr := &auth.Manager{}
	pols := []data.Policy{
		{
			Name:   "Regular user ship access",
			Effect: policy.Allow,
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"Find",
				"Open",
				"Embark",
				"Disembark",
			},
		},
		{
			Name:   "Captain privledges",
			Effect: policy.Allow,
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"Fly",
				"Navigate",
				"Curse",
			},
		},
		{
			Name:   "Secret compartment access",
			Effect: policy.Allow,
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"AccessCompartments",
			},
		},
		{
			Name:   "Healthcare access",
			Effect: policy.Allow,
			Resources: []string{
				"Healthcare",
			},
			Actions: []string{
				"PresentHMOcard",
				"WaitToSeeDoc",
				"GetMedicalAdvice",
			},
		},
	}

	req := &data.Request{
		Action:   "Fire",
		Resource: "Serenity",
		User:     "malreynolds",
	}

	//	Act
	err := mgr.DoPoliciesAllow(req, pols)

	//	Assert
	if err == nil {
		t.Errorf("DoPoliciesAllow - should implicitly deny request, but did not get error")
	}

}

func TestManager_DoPoliciesAllow_ExplicitDeny_ReturnsError(t *testing.T) {

	//	Arrange
	mgr := &auth.Manager{}
	pols := []data.Policy{
		{
			Name:   "Regular user ship access",
			Effect: policy.Allow,
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"Find",
				"Open",
				"Embark",
				"Disembark",
			},
		},
		{
			Name:   "Captain privledges",
			Effect: policy.Allow,
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"Fly",
				"Navigate",
				"Curse",
			},
		},
		{
			Name:   "Secret compartment access",
			Effect: policy.Allow,
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"AccessCompartments",
			},
		},
		{
			Name:   "Deny all ship access",
			Effect: policy.Deny, // Policy deny
			Resources: []string{
				"Serenity",
			},
			Actions: []string{
				"<.*>", // Using a regex wildcard
			},
		},
		{
			Name:   "Healthcare access",
			Effect: policy.Allow,
			Resources: []string{
				"Healthcare",
			},
			Actions: []string{
				"PresentHMOcard",
				"WaitToSeeDoc",
				"GetMedicalAdvice",
			},
		},
	}

	//	Act
	err1 := mgr.DoPoliciesAllow(
		&data.Request{
			Action:   "Open",
			Resource: "Serenity",
			User:     "malreynolds",
		}, pols)

	err2 := mgr.DoPoliciesAllow(
		&data.Request{
			Action:   "PresentHMOcard",
			Resource: "Healthcare",
			User:     "malreynolds",
		}, pols)

	//	Assert
	if err1 == nil {
		t.Errorf("DoPoliciesAllow - should explicitly deny request, but did not get error")
	}

	if err2 != nil {
		t.Errorf("DoPoliciesAllow - should allow request, but got error: %v", err2)
	}

}
