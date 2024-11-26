package users

import (
	"net/http"
	"time-tracker/internal/utils"
)

type forgotForm struct {
	Email string `form:"email" validate:"required,email"`
}

func (h *UsersHandler) HandleForgotPassword(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)
	if user != nil {
		utils.RedirectDashboard(w, r)
		return
	}

	var form forgotForm
	formErrors := utils.FormErrors{}

	if r.Method == http.MethodPost {
		err := utils.ParseFormToStruct(r, &form)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		formErrors = utils.NewValidator(&form).Validate()
		if !formErrors.HasErrors() {
			utils.RenderTemplate(w, []string{"forgot-password"}, utils.TplData{
				"Title":  "Forgot password",
				"Errors": formErrors,
				"Form":   form,
			})
		}

	}
	utils.RenderTemplate(w, []string{"forgot-password"}, utils.TplData{
		"Title":  "Forgot password",
		"Errors": formErrors,
		"Form":   form,
	})
}
