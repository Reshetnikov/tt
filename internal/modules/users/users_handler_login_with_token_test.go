//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/users --tags=unit -cover -run TestHandleLoginWithToken.*
package users

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHandleLoginWithToken_UserAlreadyLoggedIn(t *testing.T) {
	SetAppDir()
	handler := NewUsersHandlers(new(MockUsersService))

	loggedInUser := &User{ID: 1, Name: "Test User"}
	req := httptest.NewRequest(http.MethodGet, "/login-with-token?token=123", nil)
	w := httptest.NewRecorder()

	// Set the logged-in user in the context
	ctx := req.Context()
	ctx = context.WithValue(ctx, ContextUserKey, loggedInUser)
	req = req.WithContext(ctx)

	handler.HandleLoginWithToken(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusSeeOther, resp.StatusCode)
	assert.Equal(t, "/dashboard", resp.Header.Get("Location"))
}

func TestHandleLoginWithToken_MissingToken(t *testing.T) {
	SetAppDir()
	handler := NewUsersHandlers(new(MockUsersService))

	req := httptest.NewRequest(http.MethodGet, "/login-with-token", nil)
	w := httptest.NewRecorder()

	handler.HandleLoginWithToken(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, w.Body.String(), "The login link is invalid or expired")
}

func TestHandleLoginWithToken_InvalidToken(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	handler := NewUsersHandlers(mockService)

	token := "invalid_token"
	mockService.On("LoginWithToken", token).Return((*Session)(nil), errors.New("invalid token"))

	req := httptest.NewRequest(http.MethodGet, "/login-with-token?token="+token, nil)
	w := httptest.NewRecorder()

	handler.HandleLoginWithToken(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, w.Body.String(), "The login link is invalid or expired")

	mockService.AssertExpectations(t)
}

func TestHandleLoginWithToken_Success(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	handler := NewUsersHandlers(mockService)

	token := "valid_token"
	session := &Session{
		SessionID: "session_id",
		Expiry:    time.Now().Add(24 * time.Hour),
	}
	mockService.On("LoginWithToken", token).Return(session, nil)

	req := httptest.NewRequest(http.MethodGet, "/login-with-token?token="+token, nil)
	w := httptest.NewRecorder()

	handler.HandleLoginWithToken(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusSeeOther, resp.StatusCode)
	assert.Equal(t, "/dashboard", resp.Header.Get("Location"))

	cookie := resp.Cookies()[0]
	assert.Equal(t, session.SessionID, cookie.Value)
	assert.True(t, cookie.HttpOnly)

	mockService.AssertExpectations(t)
}
