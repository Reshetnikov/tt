package users

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
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
				message := fmt.Sprintf(
					`Your account is not activated. Please check your email and follow the activation link. 
					If you didn’t receive the email, <a href="/signup-success?email=%s">click here to resend it</a>.`,
					url.QueryEscape(html.EscapeString(form.Email)),
				)
				formErrors.Add("Common", message)
			} else {
				formErrors.Add("Common", utils.Ukfirst(err.Error()))
			}
		}
	}
	utils.RenderTemplate(w, []string{"login"}, utils.TplData{
		"Title":  "Log In",
		"Errors": formErrors,
		"Form":   form,
	})
}
