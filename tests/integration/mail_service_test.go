//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMailService_SendEmail_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	ms, err := NewMailService()
	assert.NoError(t, err, "failed to create MailService")

	err = ms.SendEmail(
		"verified-email@example.com",
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

	ms, err := NewMailService()
	assert.NoError(t, err, "failed to create MailService")

	err = ms.SendEmail(
		"verified-email@example.com",
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

	ms, err := NewMailService()
	assert.NoError(t, err, "failed to create MailService")

	err = ms.SendEmail(
		"verified-email@example.com",
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

	ms, err := NewMailService()
	assert.NoError(t, err, "failed to create MailService")

	err = ms.SendEmail(
		"verified-email@example.com",
		"suppressionlist@simulator.amazonses.com",
		"Test Subject",
		"Test Body",
	)
	assert.NoError(t, err, "sending to suppression list simulator failed")
}
