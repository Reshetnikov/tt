package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
	"time-tracker/internal/config"
)

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	layout := filepath.Join("web", "templates", "layout.gohtml")
	tmplPath := filepath.Join("web", "templates", tmpl+".gohtml")
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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index", map[string]interface{}{
		"Title": "Dashboard",
	})
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "signup", map[string]interface{}{
		"Title": "Sign Up",
	})
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load config: %v", err)
	}
	fmt.Println(":" + cfg.Port)

	// static
	fs := http.FileServer(http.Dir("./web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/dashboard", handler)
	//http.HandleFunc("/projects", handler)
	//http.HandleFunc("/projects/{project_id}", handler)
	http.HandleFunc("/tasks", handler)
	http.HandleFunc("/reports", handler)
	http.HandleFunc("/login", handler)
	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/profile", handler)
	http.HandleFunc("/settings", handler)

	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
