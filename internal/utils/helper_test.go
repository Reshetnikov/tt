// For all go:build
// If a function is defined in a file without a build tag, but is used in a file with a build tag, it is considered unused. Therefore, functions defined here are public.
package utils

import (
	"os"
	"testing"
)

func SetAppDir() {
	os.Chdir("/app")
}

func TShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}
}
