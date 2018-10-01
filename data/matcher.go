package data

type matcher interface {
	Matches(p Policy, haystack []string, needle string) (matches bool, error error)
}

// DefaultMatcher is the default matcher
var DefaultMatcher = NewRegexpMatcher(512)
