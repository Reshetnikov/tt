//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/users --tags=unit -cover -run TestHandleActivation.*
package users

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHandleActivation_Success(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	handler := NewUsersHandlers(mockService)

	activationHash := "valid_hash"
	session := &Session{
		SessionID: "session_id",
		Expiry:    time.Now().Add(24 * time.Hour),
	}
	mockService.On("ActivateUser", activationHash).Return(session, nil)

	req := httptest.NewRequest(http.MethodGet, "/activation?hash="+activationHash, nil)
	w := httptest.NewRecorder()

	handler.HandleActivation(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	cookie := resp.Cookies()[0]
	assert.Equal(t, session.SessionID, cookie.Value)
	assert.True(t, cookie.HttpOnly)

	mockService.AssertExpectations(t)
}

func TestHandleActivation_MissingHash(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	handler := NewUsersHandlers(mockService)

	req := httptest.NewRequest(http.MethodGet, "/activation", nil)
	w := httptest.NewRecorder()

	handler.HandleActivation(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := w.Body.String()
	expectedSubstring := "Invalid activation link"
	if !strings.Contains(body, expectedSubstring) {
		t.Errorf("expected message to contain '%s', but it didn't", expectedSubstring)
	}

	mockService.AssertNotCalled(t, "ActivateUser")
}

func TestHandleActivation_InvalidHash(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	handler := NewUsersHandlers(mockService)

	activationHash := "invalid_hash"
	mockService.On("ActivateUser", activationHash).Return(nil, errors.New("activation failed"))

	req := httptest.NewRequest(http.MethodGet, "/activation?hash="+activationHash, nil)
	w := httptest.NewRecorder()

	handler.HandleActivation(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := w.Body.String()
	expectedSubstring := "Activation Failed"
	if !strings.Contains(body, expectedSubstring) {
		t.Errorf("expected message to contain '%s', but it didn't", expectedSubstring)
	}

	mockService.AssertExpectations(t)
}

func TestHandleActivation_UserAlreadyLoggedIn(t *testing.T) {
	mockService := new(MockUsersService)
	handler := NewUsersHandlers(mockService)

	loggedInUser := &User{ID: 1, Name: "Test User"}
	req := httptest.NewRequest(http.MethodGet, "/activation", nil)
	w := httptest.NewRecorder()

	ctx := req.Context()
	ctx = context.WithValue(ctx, ContextUserKey, loggedInUser)
	req = req.WithContext(ctx)

	handler.HandleActivation(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusSeeOther, resp.StatusCode)
	assert.Equal(t, "/dashboard", resp.Header.Get("Location"))
}
