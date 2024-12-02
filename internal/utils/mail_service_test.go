// For all go:build
package utils

import (
	"testing"
	"time-tracker/internal/config"

	"github.com/stretchr/testify/require"
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
