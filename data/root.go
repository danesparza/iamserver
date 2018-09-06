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
func (store Manager) GetKey(entityType string, keyPart ...string) string {
	allparts := []string{}
	allparts = append(allparts, entityType)
	allparts = append(allparts, keyPart...)
	return strings.Join(allparts, "_")
}
