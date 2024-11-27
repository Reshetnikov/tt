package users

import (
	"fmt"
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
	timeUntilResend := 0

	if r.Method == http.MethodPost {
		err := utils.ParseFormToStruct(r, &form)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		formErrors = utils.NewValidator(&form).Validate()
		if formErrors.HasErrors() {
			renderForgotPassword(w, formErrors, form, timeUntilResend, false)
			return
		}

		timeUntilResend, err = h.usersService.SendLinkToLogin(form.Email)
		if err == nil {
			renderForgotPassword(w, formErrors, form, timeUntilResend, true)
			return
		}

		if err == ErrUserNotFound {
			formErrors.Add("Email", "Email not found")
		} else if err == ErrAccountNotActivated {
			formErrors.Add("Email", getNotActivatedMessage(form.Email))
		} else if err == ErrTimeUntilResend {
			// see ErrorMessage
		} else {
			formErrors.Add("Common", "Error. Please try again later.")
		}

	}

	renderForgotPassword(w, formErrors, form, timeUntilResend, false)
}

func renderForgotPassword(w http.ResponseWriter, formErrors utils.FormErrors, form forgotForm, timeUntilResend int, saveOk bool) {
	errorMessage := ""
	if timeUntilResend > 0 {
		errorMessage = fmt.Sprintf("Wait %d sec.", timeUntilResend)
	}
	utils.RenderTemplate(w, []string{"forgot-password"}, utils.TplData{
		"Title":           "Forgot Password?",
		"Errors":          formErrors,
		"Form":            form,
		"TimeUntilResend": timeUntilResend,
		"ErrorMessage":    errorMessage,
		"SaveOk":          saveOk,
	})
}
