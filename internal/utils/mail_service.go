package utils

import (
	"bytes"
	"context"
	"path/filepath"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

var activationTplPath = filepath.Join("web", "templates", "email", "activation.html")
var loginWithTokenTplPath = filepath.Join("web", "templates", "email", "login-with-token.html")

// For tests. Instead of *ses.Client
type SESClient interface {
	SendEmail(ctx context.Context, params *ses.SendEmailInput, optFns ...func(*ses.Options)) (*ses.SendEmailOutput, error)
	GetSendQuota(ctx context.Context, params *ses.GetSendQuotaInput, optFns ...func(*ses.Options)) (*ses.GetSendQuotaOutput, error)
}

type MailService struct {
	// client *ses.Client
	client    SESClient
	emailFrom string
}

func NewMailService(emailFrom string) (*MailService, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}
	client := ses.NewFromConfig(cfg)

	return &MailService{
		client:    client,
		emailFrom: emailFrom,
	}, nil
}

func (ms *MailService) sendEmail(to string, subject string, body string) error {

	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(body),
				},
			},
			Subject: &types.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(ms.emailFrom),
	}

	_, err := ms.client.SendEmail(context.Background(), input)
	if err != nil {
		return err
	}

	return nil
}

func (ms *MailService) Ping() error {
	ctx := context.Background()
	_, err := ms.client.GetSendQuota(ctx, &ses.GetSendQuotaInput{})
	if err != nil {
		return err
	}
	return nil
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
