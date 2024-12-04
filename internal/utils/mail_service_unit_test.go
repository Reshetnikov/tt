//go:build unit

package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSESClient struct {
	mock.Mock
}

func (m *MockSESClient) SendEmail(ctx context.Context, params *ses.SendEmailInput, optFns ...func(*ses.Options)) (*ses.SendEmailOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ses.SendEmailOutput), args.Error(1)
}

// docker exec -it tt-app-1 go test -v ./internal/utils --tags=unit -cover -run TestMailService.*
func TestMailService_NewMailService(t *testing.T) {
	emailFrom := "test@example.com"
	mailService, err := NewMailService(emailFrom)

	assert.NoError(t, err)
	assert.NotNil(t, mailService)
	assert.Equal(t, emailFrom, mailService.emailFrom)
}

func TestMailService_NewMailService_InvalidEnv(t *testing.T) {
	originalAWSProfile := os.Getenv("AWS_PROFILE")
	os.Setenv("AWS_PROFILE", "invalid-profile")
	defer func() {
		os.Setenv("AWS_PROFILE", originalAWSProfile)
	}()

	mailService, err := NewMailService("noreply@example.com")

	assert.Nil(t, mailService, "MailService should be nil on error")
	assert.Error(t, err, "Expected error due to invalid AWS environment configuration")
	assert.Contains(t, err.Error(), "failed to get shared config profile", "Error message should indicate a configuration issue")
}

func TestMailService_SendEmail_Error(t *testing.T) {
	SetAppDir()
	mockSESClient := new(MockSESClient)
	mailService := &MailService{
		client:    mockSESClient,
		emailFrom: "noreply@example.com",
	}

	mockSESClient.On(
		"SendEmail",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(&ses.SendEmailOutput{}, fmt.Errorf("mock SES error"))

	err := mailService.sendEmail("user@example.com", "Test Subject", "Test Body")
	assert.Error(t, err, "Expected error when SES returns an error")
	assert.Contains(t, err.Error(), "mock SES error", "Error message should match the mock error")

	mockSESClient.AssertExpectations(t)
}

func TestMailService_SendActivationEmail(t *testing.T) {
	SetAppDir()
	mockSESClient := new(MockSESClient)
	mailService := &MailService{
		client:    mockSESClient,
		emailFrom: "noreply@example.com",
	}

	mockSESClient.On(
		"SendEmail",
		mock.Anything,
		mock.MatchedBy(func(input *ses.SendEmailInput) bool {
			return input.Destination.ToAddresses[0] == "user@example.com" &&
				input.Message.Subject.Data != nil
		}),
		mock.Anything,
	).Return(&ses.SendEmailOutput{}, nil)

	err := mailService.SendActivationEmail("user@example.com", "John Doe", "http://activation-link")

	assert.NoError(t, err)
	mockSESClient.AssertExpectations(t)
}

func TestMailService_SendLoginWithTokenEmail(t *testing.T) {
	SetAppDir()
	mockSESClient := new(MockSESClient)
	mailService := &MailService{
		client:    mockSESClient,
		emailFrom: "noreply@example.com",
	}

	mockSESClient.On(
		"SendEmail",
		mock.Anything,
		mock.MatchedBy(func(input *ses.SendEmailInput) bool {
			return input.Destination.ToAddresses[0] == "user@example.com" &&
				input.Message.Subject.Data != nil
		}),
		mock.Anything,
	).Return(&ses.SendEmailOutput{}, nil)

	err := mailService.SendLoginWithTokenEmail("user@example.com", "John Doe", "http://login-link")

	assert.NoError(t, err)
	mockSESClient.AssertExpectations(t)
}

func TestMailService_TemplateError(t *testing.T) {
	os.Chdir("/tmp")
	mockSESClient := new(MockSESClient)
	mailService := &MailService{
		client:    mockSESClient,
		emailFrom: "noreply@example.com",
	}
	err := mailService.SendActivationEmail("user@example.com", "John Doe", "http://link")
	assert.Error(t, err)
	err = mailService.SendLoginWithTokenEmail("user@example.com", "John Doe", "http://link")
	assert.Error(t, err)

	SetAppDir()
	oldActivationTplPath := activationTplPath
	activationTplPath = filepath.Join("web", "templates", "test", "missing_emplate.html")
	err = mailService.SendActivationEmail("user@example.com", "John Doe", "http://link")
	assert.Error(t, err)
	activationTplPath = oldActivationTplPath

	oldLoginWithTokenTplPath := loginWithTokenTplPath
	loginWithTokenTplPath = filepath.Join("web", "templates", "test", "missing_emplate.html")
	err = mailService.SendLoginWithTokenEmail("user@example.com", "John Doe", "http://link")
	assert.Error(t, err)
	loginWithTokenTplPath = oldLoginWithTokenTplPath
}
