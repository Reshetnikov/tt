//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/users --tags=unit -cover -run TestHandleForgotPassword.*
package users

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleForgotPassword_UserLoggedIn(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	handler := &UsersHandler{usersService: mockService}

	loggedInUser := &User{ID: 1, Name: "Test User"}
	req := httptest.NewRequest(http.MethodGet, "/forgot-password", nil)
	w := httptest.NewRecorder()

	ctx := req.Context()
	ctx = context.WithValue(ctx, ContextUserKey, loggedInUser)
	req = req.WithContext(ctx)

	handler.HandleForgotPassword(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusSeeOther, resp.StatusCode)
	assert.Equal(t, "/dashboard", resp.Header.Get("Location"))
}

func TestHandleForgotPassword_GetRequest(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	handler := &UsersHandler{usersService: mockService}

	req := httptest.NewRequest(http.MethodGet, "/forgot-password", nil)
	w := httptest.NewRecorder()

	handler.HandleForgotPassword(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, w.Body.String(), "Forgot Password?")
}

func TestHandleForgotPassword_ValidPost(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	handler := &UsersHandler{usersService: mockService}

	email := "test@example.com"
	mockService.On("SendLinkToLogin", email).Return(0, nil)

	formData := "email=" + email
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.HandleForgotPassword(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, w.Body.String(), "Forgot Password?")

	mockService.AssertExpectations(t)
}

func TestHandleForgotPassword_EmailNotFound(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	handler := &UsersHandler{usersService: mockService}

	email := "notfound@example.com"
	mockService.On("SendLinkToLogin", email).Return(0, ErrUserNotFound)

	formData := "email=" + email
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.HandleForgotPassword(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, w.Body.String(), "Email not found")

	mockService.AssertExpectations(t)
}

func TestHandleForgotPassword_AccountNotActivated(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	handler := &UsersHandler{usersService: mockService}

	email := "inactive@example.com"
	mockService.On("SendLinkToLogin", email).Return(0, ErrAccountNotActivated)

	formData := "email=" + email
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.HandleForgotPassword(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, w.Body.String(), getNotActivatedMessage(email))

	mockService.AssertExpectations(t)
}

func TestHandleForgotPassword_TimeUntilResend(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	handler := &UsersHandler{usersService: mockService}

	email := "test@example.com"
	timeUntilResend := 30
	mockService.On("SendLinkToLogin", email).Return(timeUntilResend, ErrTimeUntilResend)

	formData := "email=" + email
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.HandleForgotPassword(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, w.Body.String(), "Please wait 30s")

	mockService.AssertExpectations(t)
}

func TestHandleForgotPassword_CommonError(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	handler := &UsersHandler{usersService: mockService}

	email := "test@example.com"
	mockService.On("SendLinkToLogin", email).Return(0, errors.New("some error"))

	formData := "email=" + email
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.HandleForgotPassword(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, w.Body.String(), "Error. Please try again later.")

	mockService.AssertExpectations(t)
}

func TestHandleForgotPassword_ParseFormError(t *testing.T) {
	handler := NewUsersHandlers(new(MockUsersService))

	req, _ := http.NewRequest(http.MethodPost, "/forgot-password", nil)

	w := httptest.NewRecorder()

	handler.HandleForgotPassword(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, w.Body.String(), "Bad Request")
}

func TestHandleForgotPassword_FormValidationErrors(t *testing.T) {
	handler := NewUsersHandlers(new(MockUsersService))

	invalidForm := "email=invalid_email"
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(invalidForm))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.HandleForgotPassword(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, w.Body.String(), "Email must be a valid email address")
}
