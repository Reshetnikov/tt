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

type MailService struct {
	client    *ses.Client
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

func (ms *MailService) SendEmail(to string, subject string, body string) error {

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

func (ms *MailService) SendActivationEmail(email, name, link string) error {
	templatePath := filepath.Join("web", "templates", "email", "activation.html")
	tmpl, err := template.ParseFiles(templatePath)
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

	return ms.SendEmail(
		email,
		"Activating an account in Time Tracker",
		bodyBuffer.String(),
	)
}
