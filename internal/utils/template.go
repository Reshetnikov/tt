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

// Example:
// {{ formatTimeRange .TimeStart .TimeEnd }}
func formatTimeRange(timeStart time.Time, timeEnd *time.Time, timezone string) string {
	const timeFormat = "15:04"
	const dateTimeFormat = "02 Jan 2006 15:04"

	effectiveEnd := timeEnd
	if effectiveEnd == nil {
		now, _ := NowWithTimezone(timezone)
		effectiveEnd = &now
	}

	duration := effectiveEnd.Sub(timeStart)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	var timeRange string

	if timeEnd == nil {
		now := time.Now()
		isToday := timeStart.Year() == now.Year() && timeStart.YearDay() == now.YearDay()
		if isToday {
			timeRange = fmt.Sprintf("%s - in progress", timeStart.Format(timeFormat))
		} else {
			timeRange = fmt.Sprintf("%s - in progress", timeStart.Format(dateTimeFormat))
		}
	} else {
		if timeStart.Truncate(24 * time.Hour).Equal(timeEnd.Truncate(24 * time.Hour)) {
			// Same date, only time shown
			timeRange = fmt.Sprintf("%s - %s", timeStart.Format(timeFormat), timeEnd.Format(timeFormat))
		} else {
			// Different dates, show date and time
			timeRange = fmt.Sprintf("%s - %s", timeStart.Format(dateTimeFormat), timeEnd.Format(dateTimeFormat))
		}
	}

	if hours > 0 || minutes > 0 {
		timeRange += fmt.Sprintf(" (%dh %dm)", hours, minutes)
	} else {
		timeRange += " (<1m)"
	}

	return timeRange
}

func createTemplate(w http.ResponseWriter, tplPaths []string) (templates *template.Template) {
	templates = template.New("").Funcs(template.FuncMap{
		"dict":            dict,
		"date":            dateFormat,
		"fileVersion":     fileVersion,
		"formatTimeRange": formatTimeRange,
	})

	components := filepath.Join("web", "templates", "components", "*")
	templates, err := templates.ParseGlob(components)
	if err != nil {
		slog.Error("RenderTemplate ParseGlob", "components", components, "err", err.Error())
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}
	for _, tplPath := range tplPaths {
		tplPath = filepath.Join("web", "templates", tplPath+".html")
		templates, err = templates.ParseFiles(tplPath)
		if err != nil {
			slog.Error("RenderTemplate ParseFiles", "tplPath", tplPath, "err", err.Error())
			http.Error(w, "Error loading template", http.StatusInternalServerError)
			return
		}
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

func RenderTemplate(w http.ResponseWriter, tplPaths []string, data TplData) {
	templates := createTemplate(w, tplPaths)

	layout := filepath.Join("web", "templates", "layout.html")
	templates, err := templates.ParseFiles(layout)
	if err != nil {
		slog.Error("RenderTemplate ParseFiles", "layout", layout, "err", err.Error())
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	executeTemplate(templates, w, "layout", data)
}

func RenderTemplateWithoutLayout(w http.ResponseWriter, tplPaths []string, tplName string, data TplData) {
	templates := createTemplate(w, tplPaths)
	executeTemplate(templates, w, tplName, data)
}

func RenderBlockNeedLogin(w http.ResponseWriter) {
	// w.WriteHeader(http.StatusUnauthorized) // status 401
	w.Write([]byte(`
		<div class="p-4 text-center">
			<p class="text-gray-700">You need to be logged in to access this feature. Please 
				<a href="/login" class="text-blue-500 hover:text-blue-700 underline font-semibold">log in</a>.
			</p>
		</div>
	`))
}
