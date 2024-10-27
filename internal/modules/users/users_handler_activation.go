package users

import (
	"net/http"
	"time-tracker/internal/utils"
)

// ActivationHandler — обработчик для активации учетной записи
func (h *UsersHandler) ActivationHandler(w http.ResponseWriter, r *http.Request) {
    activationHash := r.URL.Query().Get("hash")
    if activationHash == "" {
        utils.RenderTemplateError(w, r, "", "Invalid activation link")
        return
    }

    session, err := h.usersService.ActivateUser(activationHash)
    if err != nil {
        utils.RenderTemplateError(w, r, "Activation Failed", "Failed to activate the account. The activation link might be expired or invalid.")
        return
    }
    setSessionCookie(w, session.SessionID, session.Expiry)

    utils.RenderTemplate(w, r, "activation-success", map[string]interface{}{
        "Title": "Activation Successful - Logged In",
    })
}