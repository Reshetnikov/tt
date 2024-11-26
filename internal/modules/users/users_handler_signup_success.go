package users

import (
	"net/http"
	"time"
	"time-tracker/internal/utils"
)

func (h *UsersHandler) HandleSignupSuccess(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)
	if user != nil {
		utils.RedirectDashboard(w, r)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email not found", http.StatusNotFound)
		return
	}

	user = h.usersService.usersRepo.GetByEmail(email)
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if user.IsActive {
		utils.RedirectLogin(w, r)
		return
	}

	dt := 60 - int(time.Since(user.ActivationHashDate).Seconds())
	if dt < 0 {
		dt = 0
	}
	if r.Method == http.MethodPost && dt == 0 {
		h.usersService.ReSendActivationMassage(user)
	}

	utils.RenderTemplate(w, []string{"signup-success"}, utils.TplData{
		"Title": "Sign Up Successful",
		"Email": email,
		"Dt":    dt,
	})
}
