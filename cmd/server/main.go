package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
	"time-tracker/internal/config"
	"time-tracker/internal/modules/pages"
	"time-tracker/internal/modules/users"
	"time-tracker/internal/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	utils.Dump("Config", cfg)

    // Connect to the database
	db, err := connectToDatabase(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	// usersRepo := users.NewUsersRepositoryMem()
	usersRepo := users.NewUsersRepositoryPostgres(db)
	// sessionsRepo := users.NewSessionsRepositoryMem()
	sessionsRepo := users.NewSessionsRepositoryRedis(cfg.RedisAddr, "", 0)
	usersService := users.NewUsersService(usersRepo, sessionsRepo)
	usersHandlers := users.NewUsersHandlers(usersService)

	mux := http.NewServeMux()

	// static
	fsPublic := http.FileServer(http.Dir("./web/public"))
	mux.Handle("/img/", fsPublic)
	mux.Handle("/css/", fsPublic)
	mux.Handle("/favicon.ico", fsPublic)

	mux.HandleFunc("/{$}", pages.IndexHandler)
	mux.HandleFunc("/signup", usersHandlers.SignupHandler)
	mux.HandleFunc("/login", usersHandlers.LoginHandler)
	mux.HandleFunc("/activation", usersHandlers.ActivationHandler)
	mux.HandleFunc("POST /logout", usersHandlers.LogoutHandler)

	mux.HandleFunc("/dashboard", pages.IndexHandler)
	// mux.HandleFunc("/projects", handler)
	// http.HandleFunc("/projects/{project_id}", handler)
	// mux.HandleFunc("/tasks", pages.IndexHandler)
	// mux.HandleFunc("/reports", pages.IndexHandler)
	// mux.HandleFunc("/profile", pages.IndexHandler)
	// mux.HandleFunc("/settings", pages.IndexHandler)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})


	muxSession := usersService.SessionMiddleware(mux)

	log.Fatal(http.ListenAndServe(":8080", muxSession))
}

func connectToDatabase(cfg *config.Config) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(cfg.GetPostgresDSN())
	if err != nil {
		return nil, fmt.Errorf("error parsing database config: %w", err)
	}

	config.MaxConns = 10
	config.MaxConnLifetime = time.Hour

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("error pinging database: %w", err)
	}
	fmt.Println("Successfully connected to the database")
	return db, nil
}