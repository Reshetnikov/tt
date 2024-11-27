package users

import (
	"net/http"
	"time-tracker/internal/utils"
)

// /login-with-token?token=123
func (h *UsersHandler) HandleLoginWithToken(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)
	if user != nil {
		utils.RedirectDashboard(w, r)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		utils.RenderTemplate(w, []string{"error"}, utils.TplData{
			"Title":   "Trouble Logging In?",
			"Message": "The login link is invalid or expired. <a href=\"/forgot-password\" class=\"text-blue-500 hover:underline\">Please request a new one</a>.",
		})
		return
	}

	session, err := h.usersService.LoginWithToken(token)
	if err != nil {
		utils.RenderTemplate(w, []string{"error"}, utils.TplData{
			"Title":   "Trouble Logging In?",
			"Message": "The login link is invalid or expired. <a href=\"/forgot-password\" class=\"text-blue-500 hover:underline\">Please request a new one</a>.",
		})
		return
	}
	setSessionCookie(w, session.SessionID, session.Expiry)
	utils.RedirectDashboard(w, r)
}
