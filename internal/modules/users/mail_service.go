package users

type MailService interface {
	SendActivationEmail(email, name, link string) error
	SendLoginWithTokenEmail(email, name, link string) error
}
