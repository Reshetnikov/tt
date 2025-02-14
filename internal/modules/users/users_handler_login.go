package users

import (
	"net/http"
	"time-tracker/internal/utils"
)

type loginForm struct {
	Email    string `form:"email" validate:"required,email"`
	Password string `form:"password" validate:"required"`
}

func (h *UsersHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)
	if user != nil {
		utils.RedirectDashboard(w, r)
		return
	}

	var form loginForm
	formErrors := utils.FormErrors{}

	if r.Method == http.MethodPost {
		err := utils.ParseFormToStruct(r, &form)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		formErrors = utils.NewValidator(&form).Validate()
		if formErrors.HasErrors() {
			renderLogin(w, formErrors, form)
			return
		}

		var session *Session
		session, err = h.usersService.LoginUser(form.Email, form.Password)
		if err == nil {
			setSessionCookie(w, session.SessionID, session.Expiry)
			utils.RedirectDashboard(w, r)
			return
		}

		if err == ErrInvalidEmailOrPassword {
			formErrors.Add("Common", "Invalid email or password")
		} else if err == ErrAccountNotActivated {
			formErrors.Add("Common", getNotActivatedMessage(form.Email))
		} else {
			formErrors.Add("Common", "Error. Please try again later.")
		}

	}

	renderLogin(w, formErrors, form)
}
func renderLogin(w http.ResponseWriter, formErrors utils.FormErrors, form loginForm) {
	utils.RenderTemplate(w, []string{"login"}, utils.TplData{
		"Title":  "Log In",
		"Errors": formErrors,
		"Form":   form,
	})
}
