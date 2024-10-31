package users

import (
	"fmt"
	"net/http"
	"time-tracker/internal/utils"
)

type signupForm struct {
	Name                 string `form:"name" validate:"required,min=2,max=40"`
	Email                string `form:"email" validate:"required,email"`
	Password             string `form:"password" validate:"required,min=8"`
	PasswordConfirmation string `form:"password_confirmation" validate:"required,eqfield=Password" label:"Confirm Password"`
}

func (h *UsersHandler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Printf("-----%+v\n", r)
	var form signupForm
	var formErrors utils.FormErrors 
	if r.Method == http.MethodPost {
		
		err := utils.ParseFormToStruct(r, &form)
		if err == nil {

			formErrors = utils.NewValidator(&form).Validate()
			if formErrors == nil {

				err = h.usersService.RegisterUser(RegisterUserData{
					Name:                 form.Name,
					Email:                form.Email,
					Password:             form.Password,
				})
				if err == nil {
					println("Redirect /signup-success")
					// http.Redirect(w, r, "/signup-success", http.StatusSeeOther)
					utils.RenderTemplate(w, r, "signup-success", map[string]interface{}{
						"Title":  "Sign Up Successful",
					})
					return
				} else {
					fmt.Printf("SignupHandler()-----%+v\n", err)
					formErrors, _ = err.(utils.FormErrors)
				}
			}
		}

		if formErrors == nil && err != nil {
			utils.RenderTemplateError(w, r, "Signup Failed", err.Error())
			return
		}
	}
	utils.RenderTemplate(w, r, "signup", map[string]interface{}{
		"Title":  "Sign Up",
		"Errors": formErrors,
		"Form":   form,
	})
}