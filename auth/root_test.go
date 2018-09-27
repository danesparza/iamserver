package auth_test

import (
	"os"
	"path"
)

//	Gets the database path for this environment:
func getTestFiles() (string, string) {
	systemdb := ""
	tokendb := ""

	testRoot := os.Getenv("IAM_TEST_ROOT")

	if testRoot != "" {
		systemdb = path.Join(testRoot, "system")
		tokendb = path.Join(testRoot, "token")
	}

	return systemdb, tokendb
}
