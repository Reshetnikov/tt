//go:build unit

// docker exec -it tt-app-1 go test -v ./cmd/server --tags=unit -cover -run TestMain_.*
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time-tracker/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain_ConnectToDatabase(t *testing.T) {
	// t.Run("TestConnectToDatabaseSuccess", func(t *testing.T) {
	// 	cfg := &config.Config{
	// 		DBHost:     "localhost",
	// 		DBPort:     "5432",
	// 		DBUser:     "user",
	// 		DBPassword: "password",
	// 		DBName:     "testdb",
	// 		DBSSLMode:  "require",
	// 	}
	// 	// Mock database connection logic or ensure a real database is running
	// 	db, err := connectToDatabase(cfg)
	// 	require.NoError(t, err)
	// 	assert.NotNil(t, db)
	// 	defer db.Close()
	// })

	t.Run("TestConnectToDatabaseFailure", func(t *testing.T) {
		cfg := &config.Config{
			DBHost:     "invalidhost",
			DBPort:     "5432",
			DBUser:     "user",
			DBPassword: "password",
			DBName:     "testdb",
			DBSSLMode:  "require",
		}
		// Expect failure due to invalid host
		_, err := connectToDatabase(cfg)
		require.Error(t, err)
	})
}

func TestMain_ConnectToRedis(t *testing.T) {
	// t.Run("TestConnectToRedisSuccess", func(t *testing.T) {
	// 	cfg := &config.Config{
	// 		RedisAddr: "localhost:6379",
	// 	}
	// 	// Mock Redis connection logic or ensure Redis is running
	// 	redisClient, err := connectToRedis(cfg)
	// 	require.NoError(t, err)
	// 	assert.NotNil(t, redisClient)
	// })

	t.Run("TestConnectToRedisFailure", func(t *testing.T) {
		cfg := &config.Config{
			RedisAddr: "invalidhost:6379",
		}
		// Expect failure due to invalid address
		_, err := connectToRedis(cfg)
		require.Error(t, err)
	})
}

func TestMain_RecoveryMiddleware(t *testing.T) {
	t.Run("TestRecoveryMiddlewareNoPanic", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello, World"))
		})
		recovery := recoveryMiddleware(handler)
		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		// Capture the response
		rr := httptest.NewRecorder()
		recovery.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "Hello, World")
	})

	t.Run("TestRecoveryMiddlewareWithPanic", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("Something went wrong")
		})
		recovery := recoveryMiddleware(handler)
		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		// Capture the response
		rr := httptest.NewRecorder()
		recovery.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadGateway, rr.Code)
		assert.Contains(t, rr.Body.String(), http.StatusText(http.StatusBadGateway))
	})
}

func TestMain_SetLogger(t *testing.T) {
	t.Run("TestSetLoggerDevelopment", func(t *testing.T) {
		os.Setenv("APP_ENV", "development")
		setLogger("development")
		// Check if logger is set to development log handler
		// In a real test, you would mock or check for log output if necessary
		assert.NotNil(t, slog.Default())
	})

	t.Run("TestSetLoggerProduction", func(t *testing.T) {
		os.Setenv("APP_ENV", "production")
		setLogger("production")
		// Check if logger is set to production log handler
		// In a real test, you would mock or check for log output if necessary
		assert.NotNil(t, slog.Default())
	})
}

func TestMain_MainHandler(t *testing.T) {
	t.Run("TestMainHandler", func(t *testing.T) {
		// Simulate application startup
		cfg := config.LoadConfig()
		assert.NotNil(t, cfg)

		// Test handlers can be attached and serve basic HTTP requests
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello, world!")
		})

		server := &http.Server{
			Addr:    ":8088",
			Handler: mux,
		}

		go func() {
			err := server.ListenAndServe()
			require.NoError(t, err)
		}()
		// Ensure server is running and responding
		resp, err := http.Get("http://localhost:8088/")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
