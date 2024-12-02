//go:build manual

package email

import (
	"flag"
	"testing"
)

var testEmail string

func init() {
	flag.StringVar(&testEmail, "email", "", "Email address to send registration email")
}

// docker exec -it tt-app-1 go test -v ./tests/email/registration_test.go -email=my_mail@gmail.com
// docker exec -it tt-app-1 go test -v ./... --tags=manual -run TestRegistration -email=my_mail@gmail.com
func TestRegistrationEmail(t *testing.T) {
	if testEmail == "" {
		t.Skip("Run with -email flag to execute")
	}
	t.Log("Test Registration Email " + testEmail)
}
