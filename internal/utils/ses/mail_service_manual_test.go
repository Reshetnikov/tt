//go:build manual

// The tests are designed for manual launch and visual control of the result.
package ses

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
)

var email = flag.String("email", "", "Email address to send registration email")
var name = flag.String("name", "", "Name address to send registration email")

func getEmailAndName(t *testing.T) (string, string) {
	if *email == "" {
		t.Skip("Run with -email flag to execute")
	}
	if *name == "" {
		*name = "Just God"
	}
	return *email, *name
}

// docker exec -it tt-app-1 go test -v ./internal/utils/ses --tags=manual -run TestMailService_SendActivationEmail_Manual -email=
func TestMailService_SendActivationEmail_Manual(t *testing.T) {
	SetAppDir()
	email, name := getEmailAndName(t)
	ms := NewMailServiceForTest(t)
	err := ms.SendActivationEmail(email, name, TestActivationURL)
	if err == nil {
		t.Log("Mail was successfully sent to " + email)
	}
	require.NoError(t, err, "failed to SendActivationEmail")
}

// docker exec -it tt-app-1 go test -v ./internal/utils/ses --tags=manual -run TestMailService_SendLoginWithTokenEmail_Manual -email=
func TestMailService_SendLoginWithTokenEmail_Manual(t *testing.T) {
	SetAppDir()
	email, name := getEmailAndName(t)
	ms := NewMailServiceForTest(t)
	err := ms.SendLoginWithTokenEmail(email, name, TestTokenURL)
	if err == nil {
		t.Log("Mail was successfully sent to " + email)
	}
	require.NoError(t, err, "failed to SendActivationEmail")
}
