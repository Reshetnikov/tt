//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/users --tags=unit -cover -run TestSetSessionCookie
package users

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSetSessionCookie(t *testing.T) {
	w := httptest.NewRecorder()

	expires := time.Now().Add(24 * time.Hour)

	setSessionCookie(w, "sessionID123", expires)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != sessionCookieName {
		t.Errorf("expected cookie name %s, got %s", sessionCookieName, cookie.Name)
	}
	if cookie.Value != "sessionID123" {
		t.Errorf("expected cookie value 'sessionID123', got %s", cookie.Value)
	}
	if cookie.Expires.Before(time.Now()) {
		t.Errorf("cookie expiration time is in the past: %v", cookie.Expires)
	}
}

// docker exec -it tt-app-1 go test -v ./internal/modules/users --tags=unit -cover -run TestGetNotActivatedMessage
func TestGetNotActivatedMessage(t *testing.T) {
	email := "test@example.com"
	expectedSubstring := `/signup-success?email=test%40example.com`

	message := getNotActivatedMessage(email)

	if !strings.Contains(message, expectedSubstring) {
		t.Errorf("expected message to contain '%s', but it didn't", expectedSubstring)
	}
}
