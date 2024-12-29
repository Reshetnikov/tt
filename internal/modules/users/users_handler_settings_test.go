//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/users --tags=unit -cover -run TestHandleSettings.*
package users

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleSettings_Success(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	mockService.On("UserUpdate", mock.Anything).Return(nil)

	formData := url.Values{
		"name":                  {"John Doe"},
		"timezone":              {"UTC"},
		"is_week_start_monday":  {"true"},
		"password":              {""},
		"password_confirmation": {""},
	}

	req, err := http.NewRequest("POST", "/settings", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	ctx := req.Context()
	ctx = context.WithValue(ctx, ContextUserKey, &User{})
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler := &UsersHandler{
		usersService: mockService,
	}

	handler.HandleSettings(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandleSettings_Unauthenticated(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)

	req, err := http.NewRequest("POST", "/settings", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := &UsersHandler{
		usersService: mockService,
	}

	handler.HandleSettings(w, req)

	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/login", w.Header().Get("Location"))
	mockService.AssertNotCalled(t, "UserUpdate")
}

func TestHandleSettings_ValidationError(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	mockService.On("UserUpdate", mock.Anything).Return(nil)

	formData := url.Values{
		"name":                  {""},
		"timezone":              {"UTC"},
		"is_week_start_monday":  {"true"},
		"password":              {""},
		"password_confirmation": {""},
	}

	req, err := http.NewRequest("POST", "/settings", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	ctx := req.Context()
	ctx = context.WithValue(ctx, ContextUserKey, &User{})
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler := &UsersHandler{
		usersService: mockService,
	}

	handler.HandleSettings(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertNotCalled(t, "UserUpdate")
}

func TestHandleSettings_FailUpdate(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	mockService.On("UserUpdate", mock.Anything).Return(assert.AnError)

	formData := url.Values{
		"name":                  {"John Doe"},
		"timezone":              {"UTC"},
		"is_week_start_monday":  {"true"},
		"password":              {""},
		"password_confirmation": {""},
	}

	req, err := http.NewRequest("POST", "/settings", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	ctx := req.Context()
	ctx = context.WithValue(ctx, ContextUserKey, &User{})
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler := &UsersHandler{
		usersService: mockService,
	}

	handler.HandleSettings(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandleSettings_ParseFormError(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	req := BadRequestPost("/settings")
	ctx := req.Context()
	ctx = context.WithValue(ctx, ContextUserKey, &User{})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler := &UsersHandler{
		usersService: mockService,
	}
	handler.HandleSettings(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleSettings_SuccessfulPasswordHashing(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	mockService.On("UserUpdate", mock.Anything).Return(nil)
	mockService.On("HashPassword", "newpassword").Return("hashed_newpassword", nil)

	formData := url.Values{
		"name":                  {"John Doe"},
		"timezone":              {"UTC"},
		"is_week_start_monday":  {"true"},
		"password":              {"newpassword"},
		"password_confirmation": {"newpassword"},
	}

	req, err := http.NewRequest("POST", "/settings", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	ctx := req.Context()
	ctx = context.WithValue(ctx, ContextUserKey, &User{})
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()

	handler := &UsersHandler{
		usersService: mockService,
	}

	handler.HandleSettings(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandleSettings_PasswordHashingError(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	mockService.On("UserUpdate", mock.Anything).Return(nil)
	mockService.On("HashPassword", "error1234").Return("", assert.AnError)

	formData := url.Values{
		"name":                  {"John Doe"},
		"timezone":              {"UTC"},
		"is_week_start_monday":  {"true"},
		"password":              {"error1234"},
		"password_confirmation": {"error1234"},
	}

	req, err := http.NewRequest("POST", "/settings", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	ctx := req.Context()
	ctx = context.WithValue(ctx, ContextUserKey, &User{})
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()

	handler := &UsersHandler{
		usersService: mockService,
	}

	handler.HandleSettings(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)
	mockService.AssertNotCalled(t, "UserUpdate")
}
