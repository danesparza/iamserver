package data

import (
	"github.com/danesparza/iamserver/policy"
	"github.com/pkg/errors"
)

// IsUserRequestAuthorized determines whether the given user is authorized to
// execute the given request
func (store Manager) IsUserRequestAuthorized(user User, request *Request) bool {
	retval := false

	//	If using the system user, the request is allowed:

	//	First, get all policies for the user
	pols, err := store.GetPoliciesForUser(user, user.Name)
	if err != nil {
		return retval
	}

	//	Next, find out if the request is authorized based on the policies
	//	that apply to the given user
	err = store.DoPoliciesAllow(request, pols)
	if err == nil {
		retval = true
	}

	return retval
}

// matcher gets the policy matcher (or gets the DefaultMatcher if one isn't specified)
func (store Manager) matcher() matcher {
	if store.Matcher == nil {
		store.Matcher = DefaultMatcher
	}
	return store.Matcher
}

// DoPoliciesAllow checks to see if the request is allowed by policy
func (store Manager) DoPoliciesAllow(r *Request, policies map[string]Policy) error {
	allowed := false
	deciders := []Policy{}

	//	Iterate through the list of policies
	for _, p := range policies {

		//	Does the action match with this policy?
		if pm, err := store.matcher().Matches(p, p.Actions, r.Action); err != nil {
			return errors.WithStack(err)
		} else if !pm {
			//	Continue to the next policy
			continue
		}

		//	Does the resource match with this policy?
		if pm, err := store.matcher().Matches(p, p.Resources, r.Resource); err != nil {
			return errors.WithStack(err)
		} else if !pm {
			//	Continue to the next policy
			continue
		}

		//	Is the policy effect 'deny'?
		//	If yes, then this overrides all allow policies.  Access is denied.
		if p.Effect != policy.Allow {
			deciders = append(deciders, p)
			return errors.WithStack(ErrRequestForcefullyDenied)
		}

		//	Policy allows access
		allowed = true
		deciders = append(deciders, p)
	}

	if !allowed {
		return errors.WithStack(ErrRequestDenied)
	}

	return nil
}
