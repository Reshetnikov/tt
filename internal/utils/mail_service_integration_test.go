//go:build integration

// docker exec -it tt-app-1 go test -v ./internal/utils --tags=integration
package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// docker exec -it tt-app-1 go test -v ./internal/utils --tags=integration -run TestMailService_SendEmail
// docker exec -it tt-app-1 go test -v ./internal/utils --tags=integration -run TestMailService_SendEmail/SuppressionList
func TestMailService_SendEmail(t *testing.T) {
	TShort(t)
	tests := []struct {
		email    string
		subject  string
		body     string
		testName string
	}{
		{SimulatorSuccess, "Test Subject", "Test Body", "Success"},
		{SimulatorBounce, "Test Subject", "Test Body", "Bounce"},
		{SimulatorComplaint, "Test Subject", "Test Body", "Complaint"},
		{SimulatorSuppressionlist, "Test Subject", "Test Body", "SuppressionList"},
	}

	ms := NewMailServiceForTest(t)

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			err := ms.sendEmail(test.email, test.subject, test.body)
			assert.NoError(t, err, "failed to send email for "+test.testName)
		})
	}
}

// docker exec -it tt-app-1 go test -v ./internal/utils --tags=integration -run TestMailService_SendActivationEmail_SimulatorSuccess
func TestMailService_SendActivationEmail_SimulatorSuccess(t *testing.T) {
	TShort(t)
	SetAppDir()
	ms := NewMailServiceForTest(t)
	err := ms.SendActivationEmail(
		SimulatorSuccess,
		"My Name",
		TestActivationURL,
	)
	assert.NoError(t, err, "failed to send email")
}

// docker exec -it tt-app-1 go test -v ./internal/utils --tags=integration -run TestMailService_SendLoginWithTokenEmail_SimulatorSuccess
func TestMailService_SendLoginWithTokenEmail_SimulatorSuccess(t *testing.T) {
	TShort(t)
	SetAppDir()
	ms := NewMailServiceForTest(t)
	err := ms.SendLoginWithTokenEmail(
		SimulatorSuccess,
		"My Name",
		TestTokenURL,
	)
	assert.NoError(t, err, "failed to send email")
}
