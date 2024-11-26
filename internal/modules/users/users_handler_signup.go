package users

import (
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"net/url"
	"time-tracker/internal/utils"
)

type signupForm struct {
	Name                 string `form:"name" validate:"required,min=2,max=40"`
	Email                string `form:"email" validate:"required,email"`
	Password             string `form:"password" validate:"required,min=8"`
	PasswordConfirmation string `form:"password_confirmation" validate:"required,eqfield=Password" label:"Confirm Password"`
	TimeZone             string `form:"timezone"`
	IsWeekStartMonday    bool   `form:"is_week_start_monday"`
}

func (h *UsersHandler) HandleSignup(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)
	if user != nil {
		utils.RedirectDashboard(w, r)
		return
	}

	var form signupForm
	formErrors := utils.FormErrors{}
	if r.Method == http.MethodPost {
		err := utils.ParseFormToStruct(r, &form)
		if err == nil {
			formErrors = utils.NewValidator(&form).Validate()
			if !formErrors.HasErrors() {
				err = h.usersService.RegisterUser(RegisterUserData{
					Name:              form.Name,
					Email:             form.Email,
					Password:          form.Password,
					TimeZone:          form.TimeZone,
					IsWeekStartMonday: form.IsWeekStartMonday,
				})
				if err == nil {
					utils.RedirectSignupSuccess(w, r, form.Email)
					return
				}
			}
		}
		if err != nil {
			if err == ErrEmailExists {
				formErrors.Add("Email", "Email is already in use")
			} else if err == ErrAccountNotActivated {
				message := fmt.Sprintf(
					`Your account is not activated. Please check your email and follow the activation link. 
					If you didn’t receive the email, <a href="/signup-success?email=%s">click here to resend it</a>.`,
					url.QueryEscape(html.EscapeString(form.Email)),
				)
				formErrors.Add("Common", message)
			} else {
				slog.Error("HandleSignup", "err", err)
				http.Error(w, "Error. Please try again later.", http.StatusBadGateway)
				return
			}
		}
	}
	utils.RenderTemplate(w, []string{"signup"}, utils.TplData{
		"Title":  "Sign Up",
		"Errors": formErrors,
		"Form":   form,
	})
}
