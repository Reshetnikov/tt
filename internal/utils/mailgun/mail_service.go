package mailgun

import (
	"bytes"
	"context"
	"path/filepath"
	"text/template"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

var activationTplPath = filepath.Join("web", "templates", "email", "activation.html")
var loginWithTokenTplPath = filepath.Join("web", "templates", "email", "login-with-token.html")

// For tests
type MailgunClient interface {
	Send(ctx context.Context, message *mailgun.Message) (string, string, error)
	GetDomain(ctx context.Context, domain string) (mailgun.DomainResponse, error)
}

type MailService struct {
	// client *mailgun.MailgunImpl
	client    MailgunClient
	emailFrom string
	domain    string
}

func NewMailService(domain, apiKey, emailFrom string) *MailService {
	mg := mailgun.NewMailgun(domain, apiKey)

	return &MailService{
		client:    mg,
		emailFrom: emailFrom,
		domain:    domain,
	}
}

func (ms *MailService) sendEmail(to string, subject string, body string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	message := mailgun.NewMessage(ms.emailFrom, subject, "", to)
	message.SetHTML(body)

	_, _, err := ms.client.Send(ctx, message)
	return err
}

func (ms *MailService) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, err := ms.client.GetDomain(ctx, ms.domain)
	return err
}

func (ms *MailService) SendActivationEmail(email, name, link string) error {
	tmpl, err := template.ParseFiles(activationTplPath)
	if err != nil {
		return err
	}

	data := struct {
		Name           string
		ActivationLink string
	}{
		Name:           name,
		ActivationLink: link,
	}

	var bodyBuffer bytes.Buffer
	if err := tmpl.Execute(&bodyBuffer, data); err != nil {
		return err
	}

	return ms.sendEmail(
		email,
		"Activating an account in Time Tracker",
		bodyBuffer.String(),
	)
}

func (ms *MailService) SendLoginWithTokenEmail(email, name, link string) error {
	tmpl, err := template.ParseFiles(loginWithTokenTplPath)
	if err != nil {
		return err
	}

	data := struct {
		Name      string
		LoginLink string
	}{
		Name:      name,
		LoginLink: link,
	}

	var bodyBuffer bytes.Buffer
	if err := tmpl.Execute(&bodyBuffer, data); err != nil {
		return err
	}

	return ms.sendEmail(
		email,
		"Login to Your Time Tracker Account",
		bodyBuffer.String(),
	)
}
