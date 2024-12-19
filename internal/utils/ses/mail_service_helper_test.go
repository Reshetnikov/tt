// For all go:build
// If a function is defined in a file without a build tag, but is used in a file with a build tag, it is considered unused. Therefore, functions defined here are public.
package ses

import (
	"testing"
	"time-tracker/internal/config"

	"github.com/stretchr/testify/require"
)

const (
	TestActivationURL        = "http://localhost:8080/activation?hash=123"
	TestTokenURL             = "http://localhost:8080/login-with-token?token=123"
	SimulatorSuccess         = "success@simulator.amazonses.com"
	SimulatorBounce          = "bounce@simulator.amazonses.com"
	SimulatorComplaint       = "complaint@simulator.amazonses.com"
	SimulatorSuppressionlist = "suppressionlist@simulator.amazonses.com"
)

func NewMailServiceForTest(t *testing.T) *MailService {
	cfg := config.LoadConfig()
	if cfg.EmailFrom == "" {
		t.Fatal("EMAIL_FROM is not set")
	}
	ms, err := NewMailService(cfg.EmailFrom)
	require.NoError(t, err, "failed to create MailService")
	return ms
}
