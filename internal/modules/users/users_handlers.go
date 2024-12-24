package users

import (
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type UsersHandler struct {
	usersService *UsersService
}

func NewUsersHandlers(usersService *UsersService) *UsersHandler {
	return &UsersHandler{
		usersService: usersService,
	}
}

func setSessionCookie(w http.ResponseWriter, sessionID string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Expires:  expires,
		HttpOnly: true,
		// Secure:   true,  // Use Secure if you work via HTTPS
		Path: "/",
	})
}

func getNotActivatedMessage(email string) string {
	return fmt.Sprintf(
		`Your account is not activated. Please check your email and follow the activation link. 
				If you didn’t receive the email, <a href="/signup-success?email=%s">click here to resend it</a>.`,
		url.QueryEscape(html.EscapeString(email)),
	)
}

var D = slog.Debug
