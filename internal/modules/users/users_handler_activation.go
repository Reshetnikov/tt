package users

import (
	"net/http"
	"time-tracker/internal/utils"
)

// HandleActivation — обработчик для активации учетной записи
func (h *UsersHandler) HandleActivation(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)
	if user != nil {
		utils.RedirectDashboard(w, r)
		return
	}

	activationHash := r.URL.Query().Get("hash")
	if activationHash == "" {
		utils.RenderTemplate(w, []string{"error"}, utils.TplData{
			"Title":   "Error",
			"Message": "Invalid activation link",
			"User":    user,
		})
		return
	}

	session, err := h.usersService.ActivateUser(activationHash)
	if err != nil {
		utils.RenderTemplate(w, []string{"error"}, utils.TplData{
			"Title":   "Activation Failed",
			"Message": "Failed to activate the account. The activation link might be expired or invalid.",
			"User":    user,
		})
		return
	}
	setSessionCookie(w, session.SessionID, session.Expiry)

	utils.RenderTemplate(w, []string{"activation-success"}, utils.TplData{
		"Title": "Activation Successful - Logged In",
		"User":  user,
	})
}
