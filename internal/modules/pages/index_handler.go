package pages

import (
	"net/http"
	"time-tracker/internal/modules/users"
	"time-tracker/internal/utils"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, []string{"index"}, utils.TplData{
		"Title": "Dashboard",
		"User":  users.GetUserFromRequest(r),
	})
}
