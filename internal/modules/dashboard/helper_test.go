// For all go:build
// If a function is defined in a file without a build tag, but is used in a file with a build tag, it is considered unused. Therefore, functions defined here are public.
package dashboard

import (
	"os"
)

func SetAppDir() {
	os.Chdir("/app")
}
