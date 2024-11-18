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
		utils.RenderTemplate(w, []string{"error"}, utils.TplData{
			"Title":   "Logout Failed",
			"Message": "Failed to log out.",
			"User":    GetUserFromRequest(r),
		})
		return
	}

	setSessionCookie(w, "", time.Unix(0, 0))

	utils.RedirectRoot(w, r)
}
