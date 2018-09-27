package auth

import "github.com/danesparza/iamserver/data"

type matcher interface {
	Matches(p data.Policy, haystack []string, needle string) (matches bool, error error)
}

// DefaultMatcher is the default matcher
var DefaultMatcher = NewRegexpMatcher(512)
