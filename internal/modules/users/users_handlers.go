package users

import (
	"net/http"
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
		Name:     "session_id",
		Value:    sessionID,
		Expires:  expires,
		HttpOnly: true,
		// Secure:   true,  // Используйте Secure, если работаете через HTTPS
		Path:     "/",
	})
}
