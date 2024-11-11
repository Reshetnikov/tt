package users

import (
	"net/http"
	"time-tracker/internal/utils"
)

// The method extracts the user from the context and adds the "User" to the template data.
func RenderTemplate(w http.ResponseWriter, r *http.Request, tplPaths []string, data utils.TplData) {
	data["User"] = GetUserFromRequest(r)
	utils.RenderTemplate(w, tplPaths, data)
}

func RenderTemplateError(w http.ResponseWriter, r *http.Request, title string, message string) {
	if title == "" {
		title = "Error"
	}
	RenderTemplate(w, r, []string{"error"}, utils.TplData{
		"Title":   title,
		"Message": message,
	})
}
