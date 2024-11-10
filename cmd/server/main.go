package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
	"time-tracker/internal/config"
	"time-tracker/internal/modules/dashboard"
	"time-tracker/internal/modules/pages"
	"time-tracker/internal/modules/users"
	"time-tracker/internal/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.LoadConfig()
	setLogger(cfg.AppEnv)
	slog.Info("======================================== Server start ========================================" /*, "Config", cfg*/)

	db, err := connectToDatabase(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	usersRepo := users.NewUsersRepositoryPostgres(db)
	sessionsRepo := users.NewSessionsRepositoryRedis(cfg.RedisAddr, "", 0)
	usersService := users.NewUsersService(usersRepo, sessionsRepo)
	usersHandlers := users.NewUsersHandlers(usersService)

	dashboardRepo := dashboard.NewDashboardRepositoryPostgres(db)
	dashboardHandler := dashboard.NewDashboardHandler(dashboardRepo)

	mux := http.NewServeMux()

	fsPublic := http.FileServer(http.Dir("./web/public"))
	mux.Handle("/img/", fsPublic)
	mux.Handle("/css/", fsPublic)
	mux.Handle("/favicon.ico", fsPublic)

	mux.HandleFunc("/{$}", pages.IndexHandler)
	mux.HandleFunc("/signup", usersHandlers.HandleSignup)
	mux.HandleFunc("/login", usersHandlers.HandleLogin)
	mux.HandleFunc("/activation", usersHandlers.HandleActivation)
	mux.HandleFunc("POST /logout", usersHandlers.HandleLogout)

	mux.HandleFunc("/dashboard", dashboardHandler.HandleDashboard)
	mux.HandleFunc("/tasks/new", dashboardHandler.HandleTasksNew)
	// mux.HandleFunc("/projects", handler)
	// http.HandleFunc("/projects/{project_id}", handler)
	// mux.HandleFunc("/tasks", pages.IndexHandler)
	// mux.HandleFunc("/reports", pages.IndexHandler)
	// mux.HandleFunc("/profile", pages.IndexHandler)
	// mux.HandleFunc("/settings", pages.IndexHandler)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	muxSession := users.SessionMiddleware(mux, sessionsRepo, usersRepo)

	server := &http.Server{
		Addr:    ":8080",
		Handler: muxSession,
	}
	log.Fatal(server.ListenAndServe())
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
	slog.Info("Successfully connected to the database")
	return db, nil
}

func setLogger(env string) {
	var handler slog.Handler
	if env == "development" {
		handler = utils.NewLogHandlerDev()
	} else {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
