package main

import (
	"fmt"
	"log"
	"net/http"
	"time-tracker/internal/config"
	"time-tracker/internal/handlers"
	"time-tracker/internal/repository"
	"time-tracker/internal/services"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load config: %v", err)
	}
	fmt.Printf("%+v\n", cfg)

	userRepo := repository.NewUserRepositoryMem()
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	mux := http.NewServeMux()

	// static
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/", handlers.IndexHandler)
	mux.HandleFunc("/dashboard", handlers.IndexHandler)
	//mux.HandleFunc("/projects", handler)
	//http.HandleFunc("/projects/{project_id}", handler)
	mux.HandleFunc("/tasks", handlers.IndexHandler)
	mux.HandleFunc("/reports", handlers.IndexHandler)
	mux.HandleFunc("/login", handlers.IndexHandler)
	mux.HandleFunc("/signup", userHandler.SignupHandler)
	mux.HandleFunc("/profile", handlers.IndexHandler)
	mux.HandleFunc("/settings", handlers.IndexHandler)

	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
