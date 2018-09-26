package auth

import (
	"github.com/danesparza/iamserver/data"
	"github.com/danesparza/iamserver/policy"
	"github.com/pkg/errors"
)

// Manager is an authorization manager
type Manager struct {
	Matcher matcher
}

// matcher gets the policy matcher (or gets the DefaultMatcher if one isn't specified)
func (a *Manager) matcher() matcher {
	if a.Matcher == nil {
		a.Matcher = DefaultMatcher
	}
	return a.Matcher
}

// DoPoliciesAllow checks to see if the request is allowed by policy
func (a *Manager) DoPoliciesAllow(r *data.Request, policies []data.Policy) error {
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
