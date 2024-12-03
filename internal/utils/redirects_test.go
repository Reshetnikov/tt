//go:build unit

package utils

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// docker exec -it tt-app-1 go test -v ./internal/utils --tags=unit -cover -run TestRedirect.*
func TestRedirectRoot(t *testing.T) {
	// Создаем тестовый ResponseRecorder
	w := httptest.NewRecorder()

	// Создаем фиктивный HTTP-запрос
	r := httptest.NewRequest(http.MethodGet, "/some-path", nil)

	// Вызываем функцию редиректа
	RedirectRoot(w, r)

	// Проверяем статус-код ответа
	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status code %d, got %d", http.StatusSeeOther, w.Code)
	}

	// Проверяем заголовок Location
	location := w.Header().Get("Location")
	if location != "/" {
		t.Errorf("Expected redirect to '/', got '%s'", location)
	}
}

func TestRedirectLogin(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/some-path", nil)

	RedirectLogin(w, r)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status code %d, got %d", http.StatusSeeOther, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/login" {
		t.Errorf("Expected redirect to '/login', got '%s'", location)
	}
}

func TestRedirectDashboard(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/some-path", nil)

	RedirectDashboard(w, r)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status code %d, got %d", http.StatusSeeOther, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/dashboard" {
		t.Errorf("Expected redirect to '/dashboard', got '%s'", location)
	}
}

func TestRedirectSignupSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/some-path", nil)

	testEmail := "test@example.com"

	RedirectSignupSuccess(w, r, testEmail)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status code %d, got %d", http.StatusSeeOther, w.Code)
	}

	expectedLocation := "/signup-success?email=" + url.QueryEscape(testEmail)
	location := w.Header().Get("Location")
	if location != expectedLocation {
		t.Errorf("Expected redirect to '%s', got '%s'", expectedLocation, location)
	}
}
