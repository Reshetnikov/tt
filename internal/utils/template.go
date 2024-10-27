package utils

import (
	"log"
	"net/http"
	"path/filepath"
	"text/template"
)

// Function for registration in the template engine
// Used to pass multiple variables from a template to a subtemplate
// Example:
// T1: {{ template "T2" dict "Label" "Name" "Type" "text" "Value" .Form.Name }}
// T2: name="{{ .Name }}" type="{{ .Type }}" value="{{ .Value }}"
func dict(values ...interface{}) map[string]interface{} {
    if len(values)%2 != 0 {
        panic("odd number of arguments in dict()")
    }
    m := make(map[string]interface{})
    for i := 0; i < len(values); i += 2 {
        key, ok := values[i].(string)
        if !ok {
            panic("non-string key in dict()")
        }
        m[key] = values[i+1]
    }
    return m
}

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
    templates := template.New("").Funcs(template.FuncMap{"dict": dict})

	layout := filepath.Join("web", "templates", "layout.html")
	tmplPath := filepath.Join("web", "templates", tmpl+".html")
	templates, err := templates.ParseFiles(layout, tmplPath)
	if err != nil {
		log.Println("Error loading template " + tmplPath + " | " + err.Error())
		http.Error(w, "Error loading template "+tmpl, http.StatusInternalServerError)
		return
	}

	components := filepath.Join("web", "templates", "components", "*")
	templates, err = templates.ParseGlob(components)
	if err != nil {
		log.Println("Error loading template " + tmplPath + " | " + err.Error())
		http.Error(w, "Error loading template "+tmpl, http.StatusInternalServerError)
		return
	}

	err = templates.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Println("Error rendering template " + tmplPath + " | " + err.Error())
		log.Println(err)
		http.Error(w, "Error rendering template "+tmpl, http.StatusInternalServerError)
	}
}

func RenderTemplateError(w http.ResponseWriter, title string, message string) {
	// http.Error(w, message, http.StatusBadRequest)
	if (title == "") {
		title = "Error"
	}
	RenderTemplate(w, "error", map[string]interface{}{
		"Title":   title,
		"Message":  message,
	})
}
