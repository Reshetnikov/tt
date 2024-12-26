//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/users --tags=unit -cover -run TestHandleSignupSuccess.*
package users

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHandleSignupSuccess_RedirectDashboard(t *testing.T) {
	SetAppDir()
	req := httptest.NewRequest(http.MethodGet, "/signup-success", nil)
	ctx := req.Context()
	ctx = context.WithValue(ctx, ContextUserKey, &User{})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	mockUsersService := &MockUsersService{}
	handler := &UsersHandler{usersService: mockUsersService}
	handler.HandleSignupSuccess(rr, req)

	require.Equal(t, http.StatusSeeOther, rr.Code)
	require.Equal(t, "/dashboard", rr.Header().Get("Location"))
}

func TestHandleSignupSuccess_EmailNotFound(t *testing.T) {
	SetAppDir()
	req := httptest.NewRequest(http.MethodGet, "/signup-success", nil)
	rr := httptest.NewRecorder()

	mockUsersService := &MockUsersService{}
	handler := &UsersHandler{usersService: mockUsersService}
	handler.HandleSignupSuccess(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
	require.Contains(t, rr.Body.String(), "Email not found")
}

func TestHandleSignupSuccess_UserNotFound(t *testing.T) {
	SetAppDir()
	req := httptest.NewRequest(http.MethodGet, "/signup-success?email=test@example.com", nil)
	rr := httptest.NewRecorder()

	mockUsersService := &MockUsersService{}
	mockUsersService.On("UserGetByEmail", "test@example.com").Return(nil)
	handler := &UsersHandler{usersService: mockUsersService}
	handler.HandleSignupSuccess(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
	require.Contains(t, rr.Body.String(), "User not found")
}

func TestHandleSignupSuccess_UserIsActive(t *testing.T) {
	SetAppDir()
	req := httptest.NewRequest(http.MethodGet, "/signup-success?email=test@example.com", nil)
	rr := httptest.NewRecorder()

	mockUsersService := &MockUsersService{}
	mockUsersService.On("UserGetByEmail", "test@example.com").Return(&User{IsActive: true})
	handler := &UsersHandler{usersService: mockUsersService}
	handler.HandleSignupSuccess(rr, req)

	require.Equal(t, http.StatusSeeOther, rr.Code)
	require.Equal(t, "/login", rr.Header().Get("Location"))
}

func TestHandleSignupSuccess_ResendConfirmation_Success(t *testing.T) {
	SetAppDir()
	req := httptest.NewRequest(http.MethodPost, "/signup-success?email=test@example.com", nil)
	rr := httptest.NewRecorder()

	user := &User{ActivationHashDate: time.Now().Add(-61 * time.Second)} // TimeUntilResend = 0
	mockUsersService := &MockUsersService{}
	mockUsersService.On("UserGetByEmail", "test@example.com").Return(user)
	mockUsersService.On("ReSendActivationEmail", user).Return(nil)

	handler := &UsersHandler{usersService: mockUsersService}
	handler.HandleSignupSuccess(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "Sign Up Successful")
	require.Contains(t, rr.Body.String(), "id=\"resend-button\"")
	require.NotContains(t, rr.Body.String(), "disabled")
	require.NotContains(t, rr.Body.String(), "Please wait")
	mockUsersService.AssertCalled(t, "ReSendActivationEmail", user)
}

func TestHandleSignupSuccess_ResendConfirmation_Wait(t *testing.T) {
	SetAppDir()
	req := httptest.NewRequest(http.MethodPost, "/signup-success?email=test@example.com", nil)
	rr := httptest.NewRecorder()

	user := &User{ActivationHashDate: time.Now().Add(-30 * time.Second)} // TimeUntilResend = 30
	mockUsersService := &MockUsersService{}
	mockUsersService.On("UserGetByEmail", "test@example.com").Return(user)

	handler := &UsersHandler{usersService: mockUsersService}
	handler.HandleSignupSuccess(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "Sign Up Successful")
	require.Contains(t, rr.Body.String(), "id=\"resend-button\"")
	require.Contains(t, rr.Body.String(), "disabled")
	require.Contains(t, rr.Body.String(), "Please wait")
	mockUsersService.AssertNotCalled(t, "ReSendActivationEmail", user)
}
