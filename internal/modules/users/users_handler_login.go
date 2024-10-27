package users

import (
	"net/http"
	"time-tracker/internal/utils"
)

type loginForm struct {
    Email    string `form:"email" validate:"required,email"`
    Password string `form:"password" validate:"required"`
}

func (h *UsersHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
    var form loginForm
    var formErrors utils.FormErrors

    if r.Method == http.MethodPost {
        err := utils.ParseFormToStruct(r, &form)
        if err == nil {
            formErrors = utils.NewValidator(&form).Validate()
            if formErrors == nil {
                session, err := h.usersService.LoginUser(form.Email, form.Password)
                if err == nil {
                    setSessionCookie(w, session.SessionID, session.Expiry)

                    http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
                    return
                } else {
                    formErrors = utils.FormErrors{"Login": {"Invalid email or password"}}
                }
            }
        }

        if formErrors == nil && err != nil {
            utils.RenderTemplateError(w, r, "Login Failed", err.Error())
            return
        }
    }

    utils.RenderTemplate(w, r, "login", map[string]interface{}{
        "Title":  "Log In",
        "Errors": formErrors,
        "Form":   form,
    })
}