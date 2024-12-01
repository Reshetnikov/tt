package utils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type MailService struct {
	client *ses.Client
}

func NewMailService() (*MailService, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("us-east-1"))
	if err != nil {
		return nil, err
	}
	client := ses.NewFromConfig(cfg)

	return &MailService{
		client: client,
	}, nil
}

func (ms *MailService) SendEmail(from string, to string, subject string, body string) error {

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
		Source: aws.String(from),
	}

	_, err := ms.client.SendEmail(context.Background(), input)
	if err != nil {
		return err
	}

	return nil
}
