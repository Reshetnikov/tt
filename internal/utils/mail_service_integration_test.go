//go:build integration

// docker exec -it tt-app-1 go test -v ./internal/utils --tags=integration
package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMailService_SendEmail_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	ms := NewMailServiceForTest(t)
	err := ms.sendEmail(
		"success@simulator.amazonses.com",
		"Test Subject",
		"Test Body",
	)
	assert.NoError(t, err, "failed to send email")
}

func TestMailService_SendEmail_Bounce(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	ms := NewMailServiceForTest(t)
	err := ms.sendEmail(
		"bounce@simulator.amazonses.com",
		"Test Subject",
		"Test Body",
	)
	assert.NoError(t, err, "sending to bounce simulator failed")
}

func TestMailService_SendEmail_Complaint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	ms := NewMailServiceForTest(t)
	err := ms.sendEmail(
		"complaint@simulator.amazonses.com",
		"Test Subject",
		"Test Body",
	)
	assert.NoError(t, err, "sending to complaint simulator failed")
}

func TestMailService_SendEmail_SuppressionList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	ms := NewMailServiceForTest(t)
	err := ms.sendEmail(
		"suppressionlist@simulator.amazonses.com",
		"Test Subject",
		"Test Body",
	)
	assert.NoError(t, err, "sending to suppression list simulator failed")
}
