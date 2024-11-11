package users

import (
	"net/http"
	"time-tracker/internal/utils"
)

type signupForm struct {
	Name                 string `form:"name" validate:"required,min=2,max=40"`
	Email                string `form:"email" validate:"required,email"`
	Password             string `form:"password" validate:"required,min=8"`
	PasswordConfirmation string `form:"password_confirmation" validate:"required,eqfield=Password" label:"Confirm Password"`
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
					Name:     form.Name,
					Email:    form.Email,
					Password: form.Password,
				})
				if err == nil {
					RenderTemplate(w, r, []string{"signup-success"}, utils.TplData{
						"Title": "Sign Up Successful",
					})
					return
				}
			}
		}
		if err != nil {
			if err == ErrEmailExists {
				formErrors.Add("Email", "Email is already in use")
			} else {
				formErrors.Add("Common", utils.Ukfirst((err.Error())))
			}
		}
	}
	RenderTemplate(w, r, []string{"signup"}, utils.TplData{
		"Title":  "Sign Up",
		"Errors": formErrors,
		"Form":   form,
	})
}
