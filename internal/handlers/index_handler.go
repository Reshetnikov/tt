package handlers

import "net/http"

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index", map[string]interface{}{
		"Title": "Dashboard",
	})
}
