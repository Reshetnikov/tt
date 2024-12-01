package users

type MailService interface {
	SendEmail(from string, to string, subject string, body string) error
}
