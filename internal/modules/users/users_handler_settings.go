package users

import (
	"log/slog"
	"net/http"
	"time-tracker/internal/utils"
)

type settingsForm struct {
	Name                 string `form:"name" validate:"required,min=2,max=40"`
	Password             string `form:"password" validate:"omitempty,min=8" label:"Change Password"`
	PasswordConfirmation string `form:"password_confirmation" validate:"omitempty,eqfield=Password" label:"Confirm Password"`
	TimeZone             string `form:"timezone" validate:"required"`
	IsWeekStartMonday    bool   `form:"is_week_start_monday"`
}

func (h *UsersHandler) HandleSettings(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)
	if user == nil {
		utils.RedirectLogin(w, r)
		return
	}

	form := settingsForm{
		Name:              user.Name,
		TimeZone:          user.TimeZone,
		IsWeekStartMonday: user.IsWeekStartMonday,
	}
	formErrors := utils.FormErrors{}

	saveOk := false
	if r.Method == http.MethodPost {
		err := utils.ParseFormToStruct(r, &form)
		D("form", "form", form)
		if err == nil {
			formErrors = utils.NewValidator(&form).Validate()
			if !formErrors.HasErrors() {
				user.Name = form.Name
				user.TimeZone = form.TimeZone
				user.IsWeekStartMonday = form.IsWeekStartMonday
				if form.Password != "" {
					hashedPassword, err := hashPassword(form.Password)
					if err != nil {
						slog.Error("HandleSettings hashPassword()", "err", err)
						http.Error(w, "Error. Please try again later.", http.StatusBadGateway)
						return
					}
					user.Password = hashedPassword
				}
				err = h.usersService.usersRepo.Update(user)
				if err != nil {
					slog.Error("HandleSettings Update()", "err", err)
					http.Error(w, "Error. Please try again later.", http.StatusBadGateway)
					return
				}
				saveOk = true
			}
		}
	}

	D("form2", "form", form)
	utils.RenderTemplate(w, []string{"settings"}, utils.TplData{
		"Title":  "Settings",
		"User":   user,
		"Errors": formErrors,
		"Form":   form,
		"SaveOk": saveOk,
	})
}
