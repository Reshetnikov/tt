package users

import (
	"net/http"
	"time"
	"time-tracker/internal/utils"
)

// "POST /logout"
func (h *UsersHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		utils.RedirectRoot(w, r)
		return
	}

	err = h.usersService.LogoutUser(cookie.Value)
	if err != nil {
		RenderTemplateError(w, r, "Logout Failed", "Failed to log out.")
		return
	}

	setSessionCookie(w, "", time.Unix(0, 0))

	utils.RedirectRoot(w, r)
}
