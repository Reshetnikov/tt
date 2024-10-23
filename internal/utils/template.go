package utils

import (
	"log"
	"net/http"
	"path/filepath"
	"text/template"
)

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	layout := filepath.Join("web", "templates", "layout.html")
	tmplPath := filepath.Join("web", "templates", tmpl+".html")
	templates, err := template.ParseFiles(layout, tmplPath)
	if err != nil {
		log.Println("Error loading template " + tmplPath)
		http.Error(w, "Error loading template "+tmpl, http.StatusInternalServerError)
		return
	}
	err = templates.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Println("Error rendering template " + tmplPath)
		log.Println(err)
		http.Error(w, "Error rendering template "+tmpl, http.StatusInternalServerError)
	}
}
