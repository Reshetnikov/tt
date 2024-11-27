package users

import (
	"fmt"
	"net/http"
	"time-tracker/internal/utils"
)

// Page /signup-success
// Button "Resend confirmation"
func (h *UsersHandler) HandleSignupSuccess(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)
	if user != nil {
		utils.RedirectDashboard(w, r)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		w.WriteHeader(http.StatusNotFound)
		utils.RenderTemplate(w, []string{"error"}, utils.TplData{
			"Title":   "Error",
			"Message": "Email not found",
		})
		return
	}

	notActiveUser := h.usersService.usersRepo.GetByEmail(email)
	if notActiveUser == nil {
		w.WriteHeader(http.StatusNotFound)
		utils.RenderTemplate(w, []string{"error"}, utils.TplData{
			"Title":   "Error",
			"Message": "User not found",
		})
		return
	}

	if notActiveUser.IsActive {
		utils.RedirectLogin(w, r)
		return
	}

	errorMessage := ""
	saveOk := false
	if r.Method == http.MethodPost {
		if timeUntilResend := notActiveUser.TimeUntilResend(); timeUntilResend == 0 {
			h.usersService.ReSendActivationEmail(notActiveUser) // will update TimeUntilResend
			saveOk = true
		} else {
			errorMessage = fmt.Sprintf("Wait %d sec.", timeUntilResend)
		}
	}

	utils.RenderTemplate(w, []string{"signup-success"}, utils.TplData{
		"Title":           "Sign Up Successful",
		"Email":           email,
		"TimeUntilResend": notActiveUser.TimeUntilResend(), // need new TimeUntilResend
		"ErrorMessage":    errorMessage,
		"SaveOk":          saveOk,
	})
}
