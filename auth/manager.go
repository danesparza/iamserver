package auth

import (
	"fmt"

	"github.com/danesparza/iamserver/data"
	"github.com/danesparza/iamserver/policy"
	"github.com/pkg/errors"
)

// Manager is an authorization manager
type Manager struct {
	DBManager data.Manager
	Matcher   matcher
}

// IsUserRequestAuthorized determines whether the given user is authorized to
// execute the given request
func (a *Manager) IsUserRequestAuthorized(user data.User, request *data.Request) (bool, error) {
	retval := false

	//	First, get all policies for the user
	pols, err := a.DBManager.GetPoliciesForUser(user, user.Name)
	if err != nil {
		return retval, fmt.Errorf("There was a problem getting policies for user: %s", err)
	}

	//	Next, find out if the request is authorized based on the policies
	//	that apply to the given user
	err = a.DoPoliciesAllow(request, pols)
	if err == nil {
		retval = true
	}

	return retval, err
}

// matcher gets the policy matcher (or gets the DefaultMatcher if one isn't specified)
func (a *Manager) matcher() matcher {
	if a.Matcher == nil {
		a.Matcher = DefaultMatcher
	}
	return a.Matcher
}

// DoPoliciesAllow checks to see if the request is allowed by policy
func (a *Manager) DoPoliciesAllow(r *data.Request, policies map[string]data.Policy) error {
	allowed := false
	deciders := []data.Policy{}

	//	Iterate through the list of policies
	for _, p := range policies {

		//	Does the action match with this policy?
		if pm, err := a.matcher().Matches(p, p.Actions, r.Action); err != nil {
			return errors.WithStack(err)
		} else if !pm {
			//	Continue to the next policy
			continue
		}

		//	Does the resource match with this policy?
		if pm, err := a.matcher().Matches(p, p.Resources, r.Resource); err != nil {
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
