//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/users --tags=unit -cover -run TestHandleLogin$
package users

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestHandleLogin(t *testing.T) {
	SetAppDir()
	tests := []struct {
		name         string
		method       string
		setupMock    func(*MockUsersService)
		rawBody      string
		formData     url.Values
		expectedCode int
		expectedPath string
		withUser     bool
		contentType  string
	}{
		{
			name:         "GET request renders login form",
			method:       http.MethodGet,
			setupMock:    func(m *MockUsersService) {},
			expectedCode: http.StatusOK,
		},
		{
			name:         "Already logged in user redirects to dashboard",
			method:       http.MethodGet,
			setupMock:    func(m *MockUsersService) {},
			withUser:     true,
			expectedCode: http.StatusSeeOther,
			expectedPath: "/dashboard",
		},
		{
			// ParseFormToStruct ParseForm err="invalid URL escape \"%%%\"
			name:         "ParseFormToStruct ParseForm err",
			method:       http.MethodPost,
			setupMock:    func(m *MockUsersService) {},
			rawBody:      "%%%",
			formData:     nil,
			contentType:  "application/x-www-form-urlencoded",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:      "Invalid form data returns validation errors",
			method:    http.MethodPost,
			setupMock: func(m *MockUsersService) {},
			formData: url.Values{
				"email":    {"invalid-email"},
				"password": {"pass"},
			},
			contentType:  "application/x-www-form-urlencoded",
			expectedCode: http.StatusOK,
		},
		{
			name:   "Successful login redirects to dashboard",
			method: http.MethodPost,
			setupMock: func(m *MockUsersService) {
				session := &Session{
					SessionID: "test-session",
					Expiry:    time.Now().Add(24 * time.Hour),
				}
				m.On("LoginUser", "test@example.com", "password123").
					Return(session, nil)
			},
			formData: url.Values{
				"email":    {"test@example.com"},
				"password": {"password123"},
			},
			contentType:  "application/x-www-form-urlencoded",
			expectedCode: http.StatusSeeOther,
			expectedPath: "/dashboard",
		},
		{
			name:   "Invalid credentials shows error message",
			method: http.MethodPost,
			setupMock: func(m *MockUsersService) {
				m.On("LoginUser", "test@example.com", "wrongpass").
					Return(nil, ErrInvalidEmailOrPassword)
			},
			formData: url.Values{
				"email":    {"test@example.com"},
				"password": {"wrongpass"},
			},
			contentType:  "application/x-www-form-urlencoded",
			expectedCode: http.StatusOK,
		},
		{
			name:   "Not activated account shows activation message",
			method: http.MethodPost,
			setupMock: func(m *MockUsersService) {
				m.On("LoginUser", "test@example.com", "password123").
					Return(nil, ErrAccountNotActivated)
			},
			formData: url.Values{
				"email":    {"test@example.com"},
				"password": {"password123"},
			},
			contentType:  "application/x-www-form-urlencoded",
			expectedCode: http.StatusOK,
		},
		{
			name:   "Server error shows generic message",
			method: http.MethodPost,
			setupMock: func(m *MockUsersService) {
				m.On("LoginUser", "test@example.com", "password123").
					Return(nil, errors.New("internal error"))
			},
			formData: url.Values{
				"email":    {"test@example.com"},
				"password": {"password123"},
			},
			contentType:  "application/x-www-form-urlencoded",
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUsersService)
			tt.setupMock(mockService)

			handler := &UsersHandler{
				usersService: mockService,
			}

			var req *http.Request
			if tt.method == http.MethodPost {
				if tt.rawBody != "" {
					req = httptest.NewRequest(tt.method, "/login", strings.NewReader(tt.rawBody))
				} else if tt.formData != nil {
					req = httptest.NewRequest(tt.method, "/login", strings.NewReader(tt.formData.Encode()))
				} else {
					req = httptest.NewRequest(tt.method, "/login", nil)
				}
				if tt.contentType != "" {
					req.Header.Set("Content-Type", tt.contentType)
				}
			} else {
				req = httptest.NewRequest(tt.method, "/login", nil)
			}

			if tt.withUser {
				ctx := req.Context()
				ctx = context.WithValue(ctx, ContextUserKey, &User{})
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			handler.HandleLogin(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("HandleLogin() status code = %v, want %v", w.Code, tt.expectedCode)
			}

			if tt.expectedPath != "" {
				location := w.Header().Get("Location")
				if location != tt.expectedPath {
					t.Errorf("HandleLogin() redirect location = %v, want %v", location, tt.expectedPath)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}
