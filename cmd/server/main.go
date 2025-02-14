package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"time"
	"time-tracker/internal/config"
	"time-tracker/internal/modules/dashboard"
	"time-tracker/internal/modules/pages"
	"time-tracker/internal/modules/users"
	"time-tracker/internal/utils"
	"time-tracker/internal/utils/mailgun"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.LoadConfig()
	setLogger(cfg.AppEnv)
	slog.Info("======================================== Server start ========================================", "Config", cfg)

	db, err := connectToDatabase(cfg)
	if err != nil {
		slog.Error("Database connection failed", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	redisClient, err := connectToRedis(cfg)
	if err != nil {
		slog.Error("Redis connection failed", "err", err)
		os.Exit(1)
	}

	// mailService, err := ses.NewMailService(cfg.EmailFrom)
	// if err != nil {
	// 	slog.Error("NewMailService failed", "err", err)
	// 	os.Exit(1)
	// }
	mailService := mailgun.NewMailService(
		cfg.MailgunDomain,
		cfg.MailgunApiKey,
		cfg.EmailFrom,
	)
	if err := mailService.Ping(); err != nil {
		slog.Error("NewMailService failed Ping", "err", err)
		os.Exit(1)
	}
	slog.Info("Successfully ping to the email service")

	// usersRepo := users.NewUsersRepositoryMem()
	usersRepo := users.NewUsersRepositoryPostgres(db)
	// sessionsRepo := users.NewSessionsRepositoryMem()
	sessionsRepo := users.NewSessionsRepositoryRedis(redisClient)
	usersService := users.NewUsersService(usersRepo, sessionsRepo, mailService, cfg.SiteUrl)
	usersHandlers := users.NewUsersHandlers(usersService)

	dashboardRepo := dashboard.NewDashboardRepositoryPostgres(db)
	dashboardHandler := dashboard.NewDashboardHandler(dashboardRepo)

	mux := http.NewServeMux()

	fsPublic := http.FileServer(http.Dir("./web/public"))
	mux.Handle("/img/", fsPublic)
	mux.Handle("/css/", fsPublic)
	mux.Handle("/js/", fsPublic)
	mux.Handle("/favicon.ico", fsPublic)

	mux.HandleFunc("/{$}", pages.IndexHandler)
	mux.HandleFunc("/signup", usersHandlers.HandleSignup)
	mux.HandleFunc("/signup-success", usersHandlers.HandleSignupSuccess)
	mux.HandleFunc("/activation", usersHandlers.HandleActivation)
	mux.HandleFunc("/login", usersHandlers.HandleLogin)
	mux.HandleFunc("/login-with-token", usersHandlers.HandleLoginWithToken)
	mux.HandleFunc("/forgot-password", usersHandlers.HandleForgotPassword)
	mux.HandleFunc("POST /logout", usersHandlers.HandleLogout)
	mux.HandleFunc("/settings", usersHandlers.HandleSettings)

	mux.HandleFunc("/dashboard", dashboardHandler.HandleDashboard)
	mux.HandleFunc("GET /tasks/new", dashboardHandler.HandleTasksNew)
	mux.HandleFunc("POST /tasks", dashboardHandler.HandleTasksCreate)
	mux.HandleFunc("GET /tasks/{id}", dashboardHandler.HandleTasksEdit)
	mux.HandleFunc("POST /tasks/{id}", dashboardHandler.HandleTasksUpdate)
	mux.HandleFunc("DELETE /tasks/{id}", dashboardHandler.HandleTasksDelete)
	mux.HandleFunc("GET /tasks", dashboardHandler.HandleTaskList)
	mux.HandleFunc("POST /tasks/update-sort-order", dashboardHandler.HandleUpdateSortOrder)
	mux.HandleFunc("/reports", dashboardHandler.HandleReports)

	mux.HandleFunc("GET /records/new", dashboardHandler.HandleRecordsNew)
	mux.HandleFunc("POST /records", dashboardHandler.HandleRecordsCreate)
	mux.HandleFunc("GET /records/{id}", dashboardHandler.HandleRecordsEdit)
	mux.HandleFunc("POST /records/{id}", dashboardHandler.HandleRecordsUpdate)
	mux.HandleFunc("DELETE /records/{id}", dashboardHandler.HandleRecordsDelete)
	mux.HandleFunc("GET /records", dashboardHandler.HandleRecordsList)
	// mux.HandleFunc("/projects", handler)
	// http.HandleFunc("/projects/{project_id}", handler)
	// mux.HandleFunc("/reports", pages.IndexHandler)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		D("main() HandleFunc(\"/\")", fmt.Sprintf("%s: %s", r.Method, r.URL.String()))
		http.NotFound(w, r)
	})

	muxSession := users.SessionMiddleware(mux, sessionsRepo, usersRepo)
	muxRecovery := recoveryMiddleware(muxSession)

	server := &http.Server{
		Addr:    ":8080",
		Handler: muxRecovery,
	}
	listenError := server.ListenAndServe()
	slog.Error("Server stop", "listenError", listenError)
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

func connectToRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: "",
		DB:       0,
	})
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	slog.Info("Successfully connected to the redis")
	return client, nil
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

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("Panic occurred", "error", err, "stack", string(debug.Stack()))
				http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

var D = slog.Debug
