//go:build manual

package utils

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var email = flag.String("email", "", "Email address to send registration email")
var name = flag.String("name", "", "Name address to send registration email")

// docker exec -it tt-app-1 go test -v ./internal/utils --tags=manual -run TestMailService_SendActivationEmail -email=
func TestMailService_SendActivationEmail(t *testing.T) {
	os.Chdir("/app")

	if *email == "" {
		t.Skip("Run with -email flag to execute")
	}
	if *name == "" {
		*name = "Just God"
	}

	ms := NewMailServiceForTest(t)
	err := ms.SendActivationEmail(*email, *name, "http://localhost:8080/activation?hash=123")
	if err == nil {
		t.Log("Mail was successfully sent to " + *email)
	}
	require.NoError(t, err, "failed to SendActivationEmail")
}
