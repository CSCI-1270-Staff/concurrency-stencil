package utils

import (
	"math/rand"
	"os"
	"testing"
)

// Mod vals by this value to prevent hardcoding tests
// + 1 is necessary because rand.Int63n(_) can return 0
var Salt int64 = rand.Int63n(1000) + 1

// GetTempDbFile creates a random file in the test's directory to be used for testing,
// returning the file's name. Once the test is done running, the file is deleted
func GetTempDbFile(t *testing.T) string {
	// file will be created in OS's default temporary directory
	tmpfile, err := os.CreateTemp("", "*.db")
	if err != nil {
		t.Fatal(err)
	}

	// Since os.CreateTemp automatically opens the file, we need to close it
	_ = tmpfile.Close()

	EnsureCleanup(t, func() {
		_ = os.Remove(tmpfile.Name())

		// Remove meta database file used for hash indices if it exists
		_ = os.Remove(tmpfile.Name() + ".meta")
	})
	return tmpfile.Name()
}
