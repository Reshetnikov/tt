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
	var form loginForm
	formErrors := utils.FormErrors{}

	if r.Method == http.MethodPost {
		err := utils.ParseFormToStruct(r, &form)
		if err == nil {
			formErrors = utils.NewValidator(&form).Validate()
			if !formErrors.HasErrors() {
				var session *Session
				session, err = h.usersService.LoginUser(form.Email, form.Password)
				if err == nil {
					setSessionCookie(w, session.SessionID, session.Expiry)

					utils.RedirectDashboard(w, r)
					return
				}
			}
		}
		if err != nil {
			if err == ErrInvalidEmailOrPassword {
				formErrors.Add("Common", "Invalid email or password")
			} else if err == ErrAccountNotActivated {
				formErrors.Add("Common", "Account not activated. Follow the link from the email to activate your account.")
			} else {
				formErrors.Add("Common", utils.Ukfirst(err.Error()))
			}
		}
	}
	RenderTemplate(w, r, "login", utils.TplData{
		"Title":  "Log In",
		"Errors": formErrors,
		"Form":   form,
	})
}
