package users

import (
	"net/http"
	"time"
	"time-tracker/internal/utils"
)

func (h *UsersHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("session_id")
    if err != nil || cookie.Value == "" {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    err = h.usersService.LogoutUser(cookie.Value)
    if err != nil {
        utils.RenderTemplateError(w, r, "Logout Failed", "Failed to log out.")
        return
    }

    http.SetCookie(w, &http.Cookie{
        Name:     "session_id",
        Value:    "",
        Path:     "/",
        Expires:  time.Unix(0, 0),
        HttpOnly: true,
        // Secure:   true, // Используйте Secure, если работаете через HTTPS
    })

    http.Redirect(w, r, "/", http.StatusSeeOther)
}