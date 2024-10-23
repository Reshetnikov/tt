package pages

import (
	"net/http"
	"time-tracker/internal/utils"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "index", map[string]interface{}{
		"Title": "Dashboard",
	})
}
