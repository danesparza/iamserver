package data

import (
	"strings"
)

// Manager is the data manager
type Manager struct {
	SystemDBpath string
	TokenDBpath  string
}

// GetKey returns a key to be used in the storage system
func (store Manager) GetKey(entityType string, keyPart ...string) []byte {
	allparts := []string{}
	allparts = append(allparts, entityType)
	allparts = append(allparts, keyPart...)
	return []byte(strings.Join(allparts, "_"))
}
