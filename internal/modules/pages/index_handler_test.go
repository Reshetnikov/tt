//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/pages --tags=unit -cover -run TestIndexHandler
package pages

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time-tracker/internal/modules/users"

	"github.com/stretchr/testify/assert"
)

func SetAppDir() {
	os.Chdir("/app")
}

func TestIndexHandler(t *testing.T) {
	SetAppDir()

	t.Run("RenderIndexPage", func(t *testing.T) {

		user := &users.User{ID: 1, Name: "Test User"}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		ctx := r.Context()
		ctx = context.WithValue(ctx, users.ContextUserKey, user)
		r = r.WithContext(ctx)

		IndexHandler(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)

		assert.Contains(t, w.Body.String(), "Dashboard")
		assert.Contains(t, w.Body.String(), "Test User")
	})
}
