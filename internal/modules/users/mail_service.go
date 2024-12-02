package users

type MailService interface {
	SendEmail(to string, subject string, body string) error
	SendActivationEmail(email, name, link string) error
}
