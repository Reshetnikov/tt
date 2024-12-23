//go:build unit

package mailgun

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMailgunClient struct {
	mock.Mock
}

func (m *MockMailgunClient) Send(ctx context.Context, message *mailgun.Message) (string, string, error) {
	args := m.Called(ctx, message)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockMailgunClient) GetDomain(ctx context.Context, domain string) (mailgun.DomainResponse, error) {
	args := m.Called(ctx, domain)
	return args.Get(0).(mailgun.DomainResponse), args.Error(1)
}

// docker exec -it tt-app-1 go test -v ./internal/utils/mailgun --tags=unit -cover -run TestMailService.*
func TestMailService_NewMailService(t *testing.T) {
	domain := "example.com"
	apiKey := "test-api-key"
	emailFrom := "test@example.com"

	mailService := NewMailService(domain, apiKey, emailFrom)

	assert.NotNil(t, mailService)
	assert.Equal(t, emailFrom, mailService.emailFrom)
	assert.Equal(t, domain, mailService.domain)
}

func TestMailService_Ping(t *testing.T) {
	mockClient := new(MockMailgunClient)

	mockClient.On(
		"GetDomain",
		mock.Anything,
		"example.com",
	).Return(mailgun.DomainResponse{}, nil)

	mailService := &MailService{
		client:    mockClient,
		emailFrom: "noreply@example.com",
		domain:    "example.com",
	}

	err := mailService.Ping()
	assert.NoError(t, err)
}

func TestMailService_Ping_Error(t *testing.T) {
	mockClient := new(MockMailgunClient)

	mockClient.On(
		"GetDomain",
		mock.Anything,
		"example.com",
	).Return(mailgun.DomainResponse{}, fmt.Errorf("mock Mailgun error"))

	mailService := &MailService{
		client:    mockClient,
		emailFrom: "noreply@example.com",
		domain:    "example.com",
	}

	err := mailService.Ping()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock Mailgun error")
}

func TestMailService_SendEmail_Error(t *testing.T) {
	mockClient := new(MockMailgunClient)
	mailService := &MailService{
		client:    mockClient,
		emailFrom: "noreply@example.com",
		domain:    "example.com",
	}

	mockClient.On(
		"Send",
		mock.Anything,
		mock.MatchedBy(func(m *mailgun.Message) bool {
			// Since we can't directly access Message fields, we'll just verify it's not nil
			// The actual message creation is handled by mailgun.NewMessage
			return m != nil
		}),
	).Return("", "", fmt.Errorf("mock Mailgun send error"))

	err := mailService.sendEmail("user@example.com", "Test Subject", "Test Body")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock Mailgun send error")
}

func TestMailService_SendActivationEmail(t *testing.T) {
	SetAppDir()
	mockClient := new(MockMailgunClient)
	mailService := &MailService{
		client:    mockClient,
		emailFrom: "noreply@example.com",
		domain:    "example.com",
	}

	mockClient.On(
		"Send",
		mock.Anything,
		mock.MatchedBy(func(m *mailgun.Message) bool {
			return m != nil
		}),
	).Return("id", "message", nil)

	err := mailService.SendActivationEmail("user@example.com", "John Doe", "http://activation-link")
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestMailService_SendLoginWithTokenEmail(t *testing.T) {
	SetAppDir()
	mockClient := new(MockMailgunClient)
	mailService := &MailService{
		client:    mockClient,
		emailFrom: "noreply@example.com",
		domain:    "example.com",
	}

	mockClient.On(
		"Send",
		mock.Anything,
		mock.MatchedBy(func(m *mailgun.Message) bool {
			return m != nil
		}),
	).Return("id", "message", nil)

	err := mailService.SendLoginWithTokenEmail("user@example.com", "John Doe", "http://login-link")
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// docker exec -it tt-app-1 go test -v ./internal/utils/mailgun --tags=unit -cover -run TestMailService_TemplateError
func TestMailService_TemplateError(t *testing.T) {
	mockClient := new(MockMailgunClient)
	mailService := &MailService{
		client:    mockClient,
		emailFrom: "noreply@example.com",
		domain:    "example.com",
	}
	mockClient.On(
		"Send",
		mock.Anything,
		mock.MatchedBy(func(m *mailgun.Message) bool {
			return m != nil
		}),
	).Return("id", "message", nil)

	os.Chdir("/tmp")
	err := mailService.SendActivationEmail("user@example.com", "John Doe", "http://link")
	assert.Error(t, err)
	err = mailService.SendLoginWithTokenEmail("user@example.com", "John Doe", "http://link")
	assert.Error(t, err)

	SetAppDir()
	oldActivationTplPath := activationTplPath
	activationTplPath = filepath.Join("web", "templates", "test", "missing_template.html")
	err = mailService.SendActivationEmail("user@example.com", "John Doe", "http://link")
	assert.Error(t, err)
	activationTplPath = oldActivationTplPath

	oldLoginWithTokenTplPath := loginWithTokenTplPath
	loginWithTokenTplPath = filepath.Join("web", "templates", "test", "missing_template.html")
	err = mailService.SendLoginWithTokenEmail("user@example.com", "John Doe", "http://link")
	assert.Error(t, err)
	loginWithTokenTplPath = oldLoginWithTokenTplPath
}
