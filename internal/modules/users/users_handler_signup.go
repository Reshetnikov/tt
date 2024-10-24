package users

import (
	"net/http"
	"time-tracker/internal/utils"
)

type SignupForm struct {
	Name                 string `form:"name" validate:"required,min=2,max=40"`
	Email                string `form:"email" validate:"required,email"`
	Password             string `form:"password" validate:"required,min=8" label1:"Password2"`
	PasswordConfirmation string `form:"password_confirmation" validate:"required,eqfield=Password" label:"Confirm Password"`
}

func (h *UsersHandler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Printf("-----%+v\n", r)
	var form SignupForm
	errors := utils.FormErrors{}
	if r.Method == http.MethodPost {
		err := utils.ParseFormToStruct(r, &form)
		if err != nil {
			http.Error(w, "Unable to parse form data", http.StatusBadRequest)
			return
		}
		errors = utils.NewValidator(&form).Validate()
		if !errors.HasErrors() {
			println("Signup.........")
		}
	}
	utils.RenderTemplate(w, "signup", map[string]interface{}{
		"Title":  "Sign Up",
		"Errors": errors,
		"Form":   form,
	})
}