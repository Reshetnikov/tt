package main

import (
	"fmt"
	"log"
	"net/http"
	"time-tracker/internal/config"
	"time-tracker/internal/modules/pages"
	"time-tracker/internal/modules/users"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load config: %v", err)
	}
	fmt.Printf("%+v\n", cfg)

	userRepo := users.NewUserRepositoryMem()
	userService := users.NewUserService(userRepo)
	userHandler := users.NewUserHandlers(userService)

	mux := http.NewServeMux()

	// static
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/", pages.IndexHandler)
	mux.HandleFunc("/dashboard", pages.IndexHandler)
	//mux.HandleFunc("/projects", handler)
	//http.HandleFunc("/projects/{project_id}", handler)
	mux.HandleFunc("/tasks", pages.IndexHandler)
	mux.HandleFunc("/reports", pages.IndexHandler)
	mux.HandleFunc("/login", pages.IndexHandler)
	mux.HandleFunc("/signup", userHandler.SignupHandler)
	mux.HandleFunc("/profile", pages.IndexHandler)
	mux.HandleFunc("/settings", pages.IndexHandler)

	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
