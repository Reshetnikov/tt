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
		RenderTemplateError(w, r, "", "Invalid activation link")
		return
	}

	session, err := h.usersService.ActivateUser(activationHash)
	if err != nil {
		RenderTemplateError(w, r, "Activation Failed", "Failed to activate the account. The activation link might be expired or invalid.")
		return
	}
	setSessionCookie(w, session.SessionID, session.Expiry)

	RenderTemplate(w, r, []string{"activation-success"}, utils.TplData{
		"Title": "Activation Successful - Logged In",
	})
}
