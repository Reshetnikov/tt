//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/users --tags=unit -cover -run TestHandleSignup_.*
package users

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandleSignup_RedirectDashboard(t *testing.T) {
	SetAppDir()
	req := httptest.NewRequest(http.MethodGet, "/signup", nil)
	ctx := req.Context()
	ctx = context.WithValue(ctx, ContextUserKey, &User{})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	mockUsersService := &MockUsersService{}
	handler := &UsersHandler{usersService: mockUsersService}
	handler.HandleSignup(rr, req)

	require.Equal(t, http.StatusSeeOther, rr.Code)
	require.Equal(t, "/dashboard", rr.Header().Get("Location"))
}

func TestHandleSignup_RenderForm_GetRequest(t *testing.T) {
	SetAppDir()
	req := httptest.NewRequest(http.MethodGet, "/signup", nil)
	rr := httptest.NewRecorder()

	mockUsersService := &MockUsersService{}
	handler := &UsersHandler{usersService: mockUsersService}
	handler.HandleSignup(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "Sign Up")
}

func TestHandleSignup_FormValidationErrors(t *testing.T) {
	SetAppDir()
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader("name=&email=invalid&password=short&password_confirmation=diff"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	mockUsersService := &MockUsersService{}
	handler := &UsersHandler{usersService: mockUsersService}
	handler.HandleSignup(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "Sign Up")
	require.Contains(t, rr.Body.String(), "Name is required")
	require.Contains(t, rr.Body.String(), "Email must be a valid email")
	require.Contains(t, rr.Body.String(), "Password must be at least 8 characters")
	require.Contains(t, rr.Body.String(), "Confirm Password must match Password")
}

func TestHandleSignup_RegisterUserSuccess(t *testing.T) {
	SetAppDir()
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader("name=Test&email=test@example.com&password=password123&password_confirmation=password123"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	mockUsersService := &MockUsersService{}
	mockUsersService.On("RegisterUser", mock.Anything).Return(nil)

	handler := &UsersHandler{usersService: mockUsersService}
	handler.HandleSignup(rr, req)

	require.Equal(t, http.StatusSeeOther, rr.Code)
	require.Equal(t, "/signup-success?email=test%40example.com", rr.Header().Get("Location"))
	mockUsersService.AssertCalled(t, "RegisterUser", RegisterUserData{
		Name:              "Test",
		Email:             "test@example.com",
		Password:          "password123",
		TimeZone:          "",
		IsWeekStartMonday: false,
	})
}

func TestHandleSignup_EmailExistsError(t *testing.T) {
	SetAppDir()
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader("name=Test&email=test@example.com&password=password123&password_confirmation=password123"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	mockUsersService := &MockUsersService{}
	mockUsersService.On("RegisterUser", mock.Anything).Return(ErrEmailExists)

	handler := &UsersHandler{usersService: mockUsersService}
	handler.HandleSignup(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "Sign Up")
	require.Contains(t, rr.Body.String(), "Email is already in use")
}

func TestHandleSignup_AccountNotActivatedError(t *testing.T) {
	SetAppDir()
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader("name=Test&email=test@example.com&password=password123&password_confirmation=password123"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	mockUsersService := &MockUsersService{}
	mockUsersService.On("RegisterUser", mock.Anything).Return(ErrAccountNotActivated)

	handler := &UsersHandler{usersService: mockUsersService}
	handler.HandleSignup(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "Sign Up")
	require.Contains(t, rr.Body.String(), getNotActivatedMessage("test@example.com"))
}

func TestHandleSignup_InternalServerError(t *testing.T) {
	SetAppDir()
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader("name=Test&email=test@example.com&password=password123&password_confirmation=password123"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	mockUsersService := &MockUsersService{}
	mockUsersService.On("RegisterUser", mock.Anything).Return(errors.New("internal error"))

	handler := &UsersHandler{usersService: mockUsersService}
	handler.HandleSignup(rr, req)

	require.Equal(t, http.StatusBadGateway, rr.Code)
	require.Contains(t, rr.Body.String(), "Error. Please try again later.")
}

func TestHandleSignup_ParseFormError(t *testing.T) {
	handler := NewUsersHandlers(new(MockUsersService))
	req := BadRequestPost("/signup")
	w := httptest.NewRecorder()
	handler.HandleSignup(w, req)
	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, w.Body.String(), "Bad Request")
}
