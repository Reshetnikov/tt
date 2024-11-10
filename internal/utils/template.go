package utils

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"
	//"time-tracker/internal/middleware"
)

type TplData map[string]interface{}

// Function for registration in the template engine
// Used to pass multiple variables from a template to a subtemplate
// Example:
// T1: {{ template "T2" dict "Label" "Name" "Type" "text" "Value" .Form.Name }}
// T2: name="{{ .Name }}" type="{{ .Type }}" value="{{ .Value }}"
func dict(values ...interface{}) TplData {
	if len(values)%2 != 0 {
		panic("odd number of arguments in dict()")
	}
	m := make(TplData)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			panic("non-string key in dict()")
		}
		m[key] = values[i+1]
	}
	return m
}

// Example:
// {{ .SelectedWeek | date "January 2, 2006" }}
func dateFormat(layout string, t time.Time) string {
	return t.Format(layout)
}

// Example:
// <link href="/css/output.css?v={{fileVersion "/css/output.css"}}" rel="stylesheet" />
func fileVersion(relPath string) string {
	absPath := filepath.Join("web/public", relPath)
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d", fileInfo.ModTime().Unix())
}

func createTemplate(w http.ResponseWriter, tplPath string) (templates *template.Template) {
	templates = template.New("").Funcs(template.FuncMap{
		"dict":        dict,
		"date":        dateFormat,
		"fileVersion": fileVersion,
	})

	components := filepath.Join("web", "templates", "components", "*")
	templates, err := templates.ParseGlob(components)
	if err != nil {
		slog.Error("RenderTemplate ParseGlob", "components", components, "err", err.Error())
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	tplPath = filepath.Join("web", "templates", tplPath+".html")
	templates, err = templates.ParseFiles(tplPath)
	if err != nil {
		slog.Error("RenderTemplate ParseFiles", "tmplPath", tplPath, "err", err.Error())
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}
	return
}

func executeTemplate(templates *template.Template, w http.ResponseWriter, tplPath string, data TplData) {
	err := templates.ExecuteTemplate(w, tplPath, data)
	if err != nil {
		slog.Error("RenderTemplate ExecuteTemplate", "err", err.Error())
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func RenderTemplate(w http.ResponseWriter, tplPath string, data TplData) {
	templates := createTemplate(w, tplPath)

	layout := filepath.Join("web", "templates", "layout.html")
	templates, err := templates.ParseFiles(layout)
	if err != nil {
		slog.Error("RenderTemplate ParseFiles", "layout", layout, "err", err.Error())
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	executeTemplate(templates, w, "layout", data)
}

func RenderTemplateWithoutLayout(w http.ResponseWriter, tplPath string, tplName string, data TplData) {
	templates := createTemplate(w, tplPath)
	executeTemplate(templates, w, tplName, data)
}
