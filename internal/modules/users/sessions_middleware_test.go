//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/users --tags=unit -cover -run TestSessionMiddleware
package users

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockSessionsRepository struct {
	mock.Mock
}

func (m *MockSessionsRepository) Create(sessionID string, session *Session) error {
	args := m.Called(sessionID, session)
	return args.Error(0)
}

func (m *MockSessionsRepository) Get(sessionID string) (*Session, error) {
	args := m.Called(sessionID)
	return args.Get(0).(*Session), args.Error(1)
}

func (m *MockSessionsRepository) Delete(sessionID string) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

type MockUsersRepository struct {
	mock.Mock
}

func (m *MockUsersRepository) Create(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUsersRepository) GetByID(id int) *User {
	args := m.Called(id)
	return args.Get(0).(*User)
}

func (m *MockUsersRepository) GetByEmail(email string) *User {
	args := m.Called(email)
	return args.Get(0).(*User)
}

func (m *MockUsersRepository) GetByActivationHash(activationHash string) *User {
	args := m.Called(activationHash)
	return args.Get(0).(*User)
}

func (m *MockUsersRepository) Update(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUsersRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestSessionMiddleware(t *testing.T) {
	mockSessionsRepo := new(MockSessionsRepository)
	mockUsersRepo := new(MockUsersRepository)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromRequest(r)
		if user == nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("No user in context"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User found in context"))
	})

	middleware := SessionMiddleware(handler, mockSessionsRepo, mockUsersRepo)

	t.Run("no cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		resp := httptest.NewRecorder()

		middleware.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("invalid session", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		sessionValue := "invalid-session"
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionValue})
		resp := httptest.NewRecorder()

		mockSessionsRepo.On("Get", sessionValue).Return((*Session)(nil), nil)

		middleware.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)
		mockSessionsRepo.AssertCalled(t, "Get", sessionValue)
	})

	t.Run("valid session but user inactive", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		sessionValue := "valid-session1"
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionValue})
		resp := httptest.NewRecorder()

		session := &Session{SessionID: sessionValue, UserID: 1, Expiry: time.Now().Add(1 * time.Hour)}
		mockSessionsRepo.On("Get", sessionValue).Return(session, nil)
		mockUsersRepo.On("GetByID", 1).Return(&User{ID: 1, IsActive: false})

		middleware.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)
		mockSessionsRepo.AssertCalled(t, "Get", sessionValue)
		mockUsersRepo.AssertCalled(t, "GetByID", 1)
	})

	t.Run("valid session and active user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		sessionValue := "valid-session2"
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionValue})
		resp := httptest.NewRecorder()

		session := &Session{SessionID: sessionValue, UserID: 2, Expiry: time.Now().Add(1 * time.Hour)}
		mockSessionsRepo.On("Get", sessionValue).Return(session, nil)
		mockUsersRepo.On("GetByID", 2).Return(&User{ID: 2, IsActive: true})

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromRequest(r)
			require.NotNil(t, user)
			require.Equal(t, 2, user.ID)
		})

		middleware := SessionMiddleware(handler, mockSessionsRepo, mockUsersRepo)
		middleware.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)
		mockSessionsRepo.AssertCalled(t, "Get", sessionValue)
		mockUsersRepo.AssertCalled(t, "GetByID", 2)
	})
}

// docker exec -it tt-app-1 go test -v ./internal/modules/users --tags=unit -cover -run TestGetUserFromRequest
func TestGetUserFromRequest(t *testing.T) {
	t.Run("user in context", func(t *testing.T) {
		user := &User{ID: 1, Name: "John Doe", IsActive: true}

		ctx := context.WithValue(context.Background(), ContextUserKey, user)
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

		retrievedUser := GetUserFromRequest(req)
		require.NotNil(t, retrievedUser)
		require.Equal(t, user, retrievedUser)
	})

	t.Run("no user in context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		retrievedUser := GetUserFromRequest(req)
		require.Nil(t, retrievedUser)
	})

	t.Run("incorrect type in context", func(t *testing.T) {
		incorrectData := "some string data"
		ctx := context.WithValue(context.Background(), ContextUserKey, incorrectData)
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

		retrievedUser := GetUserFromRequest(req)
		require.Nil(t, retrievedUser)
	})
}
