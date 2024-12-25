//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/users --tags=unit -cover -run TestHandleLogout.*
package users

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleLogout_Success(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	mockService.On("LogoutUser", "valid-session-id").Return(nil)

	req, err := http.NewRequest("POST", "/logout", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.AddCookie(&http.Cookie{
		Name:  sessionCookieName,
		Value: "valid-session-id",
	})

	rr := httptest.NewRecorder()

	handler := &UsersHandler{
		usersService: mockService,
	}

	handler.HandleLogout(rr, req)

	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Equal(t, "/", rr.Header().Get("Location"))

	mockService.AssertExpectations(t)
}

func TestHandleLogout_NoSessionCookie(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)

	req, err := http.NewRequest("POST", "/logout", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := &UsersHandler{
		usersService: mockService,
	}

	handler.HandleLogout(rr, req)

	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Equal(t, "/", rr.Header().Get("Location"))

	mockService.AssertNotCalled(t, "LogoutUser")
}

func TestHandleLogout_FailLogout(t *testing.T) {
	SetAppDir()
	mockService := new(MockUsersService)
	mockService.On("LogoutUser", "valid-session-id").Return(assert.AnError)

	req, err := http.NewRequest("POST", "/logout", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.AddCookie(&http.Cookie{
		Name:  sessionCookieName,
		Value: "valid-session-id",
	})

	rr := httptest.NewRecorder()

	handler := &UsersHandler{
		usersService: mockService,
	}

	handler.HandleLogout(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Failed to log out.")

	mockService.AssertExpectations(t)
}
