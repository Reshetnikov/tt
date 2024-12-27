//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/users --tags=unit -cover -run TestUsersService.*
package users

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const email = "test@example.com"

type MockUsersRepo struct {
	mock.Mock
}

type MockSessionsRepo struct {
	mock.Mock
}

type MockMailService struct {
	mock.Mock
}

func (m *MockUsersRepo) GetByID(id int) *User {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*User)
}
func (m *MockUsersRepo) GetByEmail(email string) *User {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*User)
}

func (m *MockUsersRepo) Create(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUsersRepo) GetByActivationHash(hash string) *User {
	args := m.Called(hash)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*User)
}

func (m *MockUsersRepo) Update(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUsersRepo) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockSessionsRepo) Create(sessionID string, session *Session) error {
	args := m.Called(sessionID, session)
	return args.Error(0)
}

func (m *MockSessionsRepo) Delete(sessionID string) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

func (m *MockSessionsRepo) Get(sessionID string) (*Session, error) {
	args := m.Called(sessionID)
	session, _ := args.Get(0).(*Session)
	return session, args.Error(1)
}

func (m *MockMailService) SendActivationEmail(email, name, link string) error {
	args := m.Called(email, name, link)
	return args.Error(0)
}

func (m *MockMailService) SendLoginWithTokenEmail(email, name, link string) error {
	args := m.Called(email, name, link)
	return args.Error(0)
}

func TestUsersService_RegisterUser(t *testing.T) {
	usersRepo := new(MockUsersRepo)
	sessionsRepo := new(MockSessionsRepo)
	mailService := new(MockMailService)
	service := NewUsersService(usersRepo, sessionsRepo, mailService, "https://example.com")

	t.Run("Success", func(t *testing.T) {
		usersRepo.On("GetByEmail", email).Return(nil).Once()
		usersRepo.On("Create", mock.Anything).Return(nil).Once()
		mailService.On("SendActivationEmail", email, "Test User", mock.Anything).Return(nil).Once()

		data := RegisterUserData{
			Name:              "Test User",
			Email:             email,
			Password:          "password123",
			TimeZone:          "UTC",
			IsWeekStartMonday: true,
		}

		err := service.RegisterUser(data)
		require.NoError(t, err)
		usersRepo.AssertExpectations(t)
		mailService.AssertExpectations(t)
	})

	t.Run("EmailExists", func(t *testing.T) {
		existingUser := &User{Email: email, IsActive: true}
		usersRepo.On("GetByEmail", email).Return(existingUser).Once()

		data := RegisterUserData{
			Name:              "Test User",
			Email:             email,
			Password:          "password123",
			TimeZone:          "UTC",
			IsWeekStartMonday: true,
		}

		err := service.RegisterUser(data)
		require.ErrorIs(t, err, ErrEmailExists)
		usersRepo.AssertExpectations(t)
	})

	t.Run("ErrAccountNotActivated", func(t *testing.T) {
		existingUser := &User{Email: email, IsActive: false}
		usersRepo.On("GetByEmail", email).Return(existingUser).Once()

		data := RegisterUserData{
			Name:              "Test User",
			Email:             email,
			Password:          "password123",
			TimeZone:          "UTC",
			IsWeekStartMonday: true,
		}

		err := service.RegisterUser(data)
		require.ErrorIs(t, err, ErrAccountNotActivated)
		usersRepo.AssertExpectations(t)
	})

	t.Run("GenerateActivationHashError", func(t *testing.T) {
		originalReader := RandomBytesReaderMock()
		defer func() { randomBytesReader = originalReader }()

		usersRepo.On("GetByEmail", email).Return(nil).Once()

		data := RegisterUserData{
			Name:              "Test User",
			Email:             email,
			Password:          "password123",
			TimeZone:          "UTC",
			IsWeekStartMonday: true,
		}

		err := service.RegisterUser(data)
		require.Error(t, err)
		usersRepo.AssertExpectations(t)
	})

	t.Run("HashPasswordError", func(t *testing.T) {
		originalFunc := BcryptGenerateFromPasswordMock()
		defer func() { bcryptGenerateFromPassword = originalFunc }()

		usersRepo.On("GetByEmail", email).Return(nil).Once()

		data := RegisterUserData{
			Name:              "Test User",
			Email:             email,
			Password:          "password123",
			TimeZone:          "UTC",
			IsWeekStartMonday: true,
		}

		err := service.RegisterUser(data)
		require.Error(t, err)
		usersRepo.AssertExpectations(t)
	})
}

func TestUsersService_ActivateUser(t *testing.T) {
	usersRepo := new(MockUsersRepo)
	sessionsRepo := new(MockSessionsRepo)
	mailService := new(MockMailService)
	service := NewUsersService(usersRepo, sessionsRepo, mailService, "https://example.com")

	t.Run("Success", func(t *testing.T) {
		user := &User{
			ID:                 1,
			ActivationHash:     "validhash",
			ActivationHashDate: time.Now().UTC(),
		}

		usersRepo.On("GetByActivationHash", "validhash").Return(user).Once()
		usersRepo.On("Update", mock.Anything).Return(nil).Once()
		sessionsRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		session, err := service.ActivateUser("validhash")
		require.NoError(t, err)
		require.NotNil(t, session)
		usersRepo.AssertExpectations(t)
		sessionsRepo.AssertExpectations(t)
	})

	t.Run("ExpiredHash", func(t *testing.T) {
		user := &User{
			ID:                 1,
			ActivationHash:     "expiredhash",
			ActivationHashDate: time.Now().Add(-16 * time.Minute),
		}

		usersRepo.On("GetByActivationHash", "expiredhash").Return(user).Once()

		session, err := service.ActivateUser("expiredhash")
		require.ErrorIs(t, err, ErrUserNotFound)
		require.Nil(t, session)
		usersRepo.AssertExpectations(t)
	})

	t.Run("UpdateError", func(t *testing.T) {
		user := &User{
			ID:                 1,
			ActivationHash:     "validhash",
			ActivationHashDate: time.Now().UTC(),
		}

		usersRepo.On("GetByActivationHash", "validhash").Return(user).Once()
		usersRepo.On("Update", mock.Anything).Return(errors.New("update error")).Once()

		session, err := service.ActivateUser("validhash")
		require.Error(t, err)
		require.Nil(t, session)
		usersRepo.AssertExpectations(t)
	})
}

func TestUsersService_LoginWithToken(t *testing.T) {
	usersRepo := new(MockUsersRepo)
	sessionsRepo := new(MockSessionsRepo)
	mailService := new(MockMailService)
	service := NewUsersService(usersRepo, sessionsRepo, mailService, "https://example.com")

	t.Run("Success", func(t *testing.T) {
		user := &User{
			ID:                 1,
			IsActive:           true,
			ActivationHash:     "validtoken",
			ActivationHashDate: time.Now().UTC(),
		}

		usersRepo.On("GetByActivationHash", "validtoken").Return(user).Once()
		usersRepo.On("Update", mock.Anything).Return(nil).Once()
		sessionsRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		session, err := service.LoginWithToken("validtoken")
		require.NoError(t, err)
		require.NotNil(t, session)
		usersRepo.AssertExpectations(t)
		sessionsRepo.AssertExpectations(t)
	})

	t.Run("ErrAccountNotActivated", func(t *testing.T) {
		user := &User{
			ID:                 1,
			IsActive:           false,
			ActivationHash:     "validtoken",
			ActivationHashDate: time.Now().UTC(),
		}

		usersRepo.On("GetByActivationHash", "validtoken").Return(user).Once()

		session, err := service.LoginWithToken("validtoken")
		require.ErrorIs(t, err, ErrAccountNotActivated)
		require.Nil(t, session)
		usersRepo.AssertExpectations(t)
	})

	t.Run("ExpiredHash", func(t *testing.T) {
		user := &User{
			ID:                 1,
			IsActive:           true,
			ActivationHash:     "validtoken",
			ActivationHashDate: time.Now().Add(-16 * time.Minute),
		}

		usersRepo.On("GetByActivationHash", "validtoken").Return(user).Once()

		session, err := service.LoginWithToken("validtoken")
		require.ErrorIs(t, err, ErrUserNotFound)
		require.Nil(t, session)
		usersRepo.AssertExpectations(t)
	})

	t.Run("UpdateError", func(t *testing.T) {
		user := &User{
			ID:                 1,
			IsActive:           true,
			ActivationHash:     "validtoken",
			ActivationHashDate: time.Now().UTC(),
		}

		usersRepo.On("GetByActivationHash", "validtoken").Return(user).Once()
		usersRepo.On("Update", mock.Anything).Return(errors.New("update error")).Once()

		session, err := service.LoginWithToken("validtoken")
		require.Error(t, err)
		require.Nil(t, session)
		usersRepo.AssertExpectations(t)
	})
}

func TestUsersService_LoginUser(t *testing.T) {
	usersRepo := new(MockUsersRepo)
	sessionsRepo := new(MockSessionsRepo)
	mailService := new(MockMailService)
	service := NewUsersService(usersRepo, sessionsRepo, mailService, "https://example.com")

	t.Run("Success", func(t *testing.T) {
		pas := "password123"
		hash, _ := service.HashPassword("password123")
		user := &User{
			Email:    email,
			Password: hash,
			IsActive: true,
		}

		usersRepo.On("GetByEmail", email).Return(user).Once()
		sessionsRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		session, err := service.LoginUser(email, pas)
		require.NoError(t, err)
		require.NotNil(t, session)

		usersRepo.AssertExpectations(t)
		sessionsRepo.AssertExpectations(t)
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		user := &User{
			Email:    email,
			Password: "$2a$10$invalidhash",
			IsActive: true,
		}

		usersRepo.On("GetByEmail", email).Return(user).Once()

		session, err := service.LoginUser(email, "wrongpassword")
		require.ErrorIs(t, err, ErrInvalidEmailOrPassword)
		require.Nil(t, session)

		usersRepo.AssertExpectations(t)
	})

	t.Run("AccountNotActivated", func(t *testing.T) {
		pas := "password123"
		hash, _ := service.HashPassword("password123")
		user := &User{
			Email:    email,
			Password: hash,
			IsActive: false,
		}

		usersRepo.On("GetByEmail", email).Return(user).Once()

		session, err := service.LoginUser(email, pas)
		require.ErrorIs(t, err, ErrAccountNotActivated)
		require.Nil(t, session)

		usersRepo.AssertExpectations(t)
	})
}

func TestUsersService_SendLinkToLogin(t *testing.T) {
	usersRepo := new(MockUsersRepo)
	mailService := new(MockMailService)
	sessionsRepo := new(MockSessionsRepo)
	service := NewUsersService(usersRepo, sessionsRepo, mailService, "https://example.com")

	t.Run("user not found", func(t *testing.T) {
		usersRepo.On("GetByEmail", "nonexistent@example.com").Return(nil).Once()

		timeUntilResend, err := service.SendLinkToLogin("nonexistent@example.com")
		require.ErrorIs(t, err, ErrUserNotFound)
		require.Equal(t, 0, timeUntilResend)
		usersRepo.AssertExpectations(t)
	})

	t.Run("user not active", func(t *testing.T) {
		user := &User{
			Email:    "inactive@example.com",
			IsActive: false,
		}
		usersRepo.On("GetByEmail", "inactive@example.com").Return(user).Once()

		timeUntilResend, err := service.SendLinkToLogin("inactive@example.com")
		require.ErrorIs(t, err, ErrAccountNotActivated)
		require.Equal(t, 0, timeUntilResend)
		usersRepo.AssertExpectations(t)
	})

	t.Run("time until resend", func(t *testing.T) {
		user := &User{
			Email:              "recent@example.com",
			IsActive:           true,
			ActivationHashDate: time.Now().Add(-55 * time.Second),
		}
		usersRepo.On("GetByEmail", "recent@example.com").Return(user).Once()

		timeUntilResend, err := service.SendLinkToLogin("recent@example.com")
		require.ErrorIs(t, err, ErrTimeUntilResend)
		require.Equal(t, 5, timeUntilResend)
		usersRepo.AssertExpectations(t)
	})

	t.Run("generate activation hash error", func(t *testing.T) {
		user := &User{
			Email:    "hashfail@example.com",
			IsActive: true,
		}
		usersRepo.On("GetByEmail", "hashfail@example.com").Return(user).Once()

		originalReader := RandomBytesReaderMock()
		defer func() { randomBytesReader = originalReader }()

		timeUntilResend, err := service.SendLinkToLogin("hashfail@example.com")
		require.Error(t, err)
		require.Equal(t, 0, timeUntilResend)
		usersRepo.AssertExpectations(t)
	})

	t.Run("update user error", func(t *testing.T) {
		user := &User{
			Email:    "updatefail@example.com",
			IsActive: true,
		}
		usersRepo.On("GetByEmail", "updatefail@example.com").Return(user).Once()
		usersRepo.On("Update", mock.MatchedBy(func(user *User) bool {
			return user.Email == "updatefail@example.com"
		})).Return(fmt.Errorf("mock update error"))

		timeUntilResend, err := service.SendLinkToLogin("updatefail@example.com")
		require.Error(t, err)
		require.EqualError(t, err, "mock update error")
		require.Equal(t, 0, timeUntilResend)
		usersRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		user := &User{
			Email:    "success@example.com",
			Name:     "Test User",
			IsActive: true,
		}
		usersRepo.On("GetByEmail", "success@example.com").Return(user).Once()
		usersRepo.On("Update", mock.MatchedBy(func(user *User) bool {
			return user.Email == "success@example.com"
		})).Return(nil)
		mailService.On("SendLoginWithTokenEmail", "success@example.com", "Test User", mock.Anything).Return(nil).Once()

		timeUntilResend, err := service.SendLinkToLogin("success@example.com")
		require.NoError(t, err)
		require.Equal(t, 60, timeUntilResend)
		usersRepo.AssertExpectations(t)
		mailService.AssertExpectations(t)
	})
}

func TestUsersService_LogoutUser(t *testing.T) {
	sessionsRepo := new(MockSessionsRepo)
	service := NewUsersService(nil, sessionsRepo, nil, "https://example.com")

	t.Run("Success", func(t *testing.T) {
		sessionsRepo.On("Delete", "session123").Return(nil).Once()
		err := service.LogoutUser("session123")
		require.NoError(t, err)
		sessionsRepo.AssertExpectations(t)
	})

	t.Run("Failure", func(t *testing.T) {
		sessionsRepo.On("Delete", "session123").Return(errors.New("delete error")).Once()
		err := service.LogoutUser("session123")
		require.Error(t, err)
		require.EqualError(t, err, "delete error")
		sessionsRepo.AssertExpectations(t)
	})
}

func TestUsersService_ReSendActivationEmail(t *testing.T) {
	usersRepo := new(MockUsersRepo)
	mailService := new(MockMailService)
	service := NewUsersService(usersRepo, nil, mailService, "https://example.com")

	user := &User{
		Email:    email,
		Name:     "Test User",
		IsActive: false,
	}

	t.Run("Success", func(t *testing.T) {
		usersRepo.On("Update", mock.Anything).Return(nil).Once()
		mailService.On("SendActivationEmail", email, "Test User", mock.Anything).Return(nil).Once()

		err := service.ReSendActivationEmail(user)
		require.NoError(t, err)
		usersRepo.AssertExpectations(t)
		mailService.AssertExpectations(t)
	})

	t.Run("UpdateError", func(t *testing.T) {
		usersRepo.On("Update", mock.Anything).Return(errors.New("update error")).Once()

		err := service.ReSendActivationEmail(user)
		require.Error(t, err)
		require.EqualError(t, err, "update error")
		usersRepo.AssertExpectations(t)
	})

	t.Run("GenerateActivationHashError", func(t *testing.T) {
		originalReader := RandomBytesReaderMock()
		defer func() { randomBytesReader = originalReader }()

		err := service.ReSendActivationEmail(user)
		require.Error(t, err)
	})
}

func TestUsersService_UserGetByEmail(t *testing.T) {
	usersRepo := new(MockUsersRepo)
	service := NewUsersService(usersRepo, nil, nil, "https://example.com")
	t.Run("Success", func(t *testing.T) {
		user := &User{
			Email: email,
		}
		usersRepo.On("GetByEmail", email).Return(user).Once()
		result := service.UserGetByEmail(email)
		require.NotNil(t, result)
		require.Equal(t, email, result.Email)
		usersRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		usersRepo.On("GetByEmail", email).Return(nil).Once()
		result := service.UserGetByEmail(email)
		require.Nil(t, result)
		usersRepo.AssertExpectations(t)
	})
}

func TestUsersService_UserUpdate(t *testing.T) {
	usersRepo := new(MockUsersRepo)
	service := NewUsersService(usersRepo, nil, nil, "https://example.com")
	user := &User{
		ID:    1,
		Email: email,
	}

	t.Run("Success", func(t *testing.T) {
		usersRepo.On("Update", user).Return(nil).Once()

		err := service.UserUpdate(user)

		require.NoError(t, err)
		usersRepo.AssertExpectations(t)
	})

	t.Run("Failure", func(t *testing.T) {
		usersRepo.On("Update", user).Return(errors.New("update error")).Once()

		err := service.UserUpdate(user)

		require.Error(t, err)
		require.EqualError(t, err, "update error")
		usersRepo.AssertExpectations(t)
	})
}

func TestUsersService_LoginWithTokenLink(t *testing.T) {
	service := NewUsersService(nil, nil, nil, "https://example.com")

	t.Run("Success", func(t *testing.T) {
		token := "validToken123"
		expectedLink := "https://example.com/login-with-token?token=" + token

		result := service.loginWithTokenLink(token)
		require.Equal(t, expectedLink, result)
	})

	t.Run("EmptyToken", func(t *testing.T) {
		token := ""
		expectedLink := "https://example.com/login-with-token?token="

		result := service.loginWithTokenLink(token)
		require.Equal(t, expectedLink, result)
	})
}

func TestUsersService_GenerateActivationHash(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		hash, err := generateActivationHash(email)
		require.NoError(t, err)
		require.NotEmpty(t, hash)
	})

	t.Run("Randomness", func(t *testing.T) {
		hash1, err1 := generateActivationHash(email)
		require.NoError(t, err1)
		require.NotEmpty(t, hash1)

		hash2, err2 := generateActivationHash(email)
		require.NoError(t, err2)
		require.NotEmpty(t, hash2)

		require.NotEqual(t, hash1, hash2, "Hashes should be unique for the same email due to randomness")
	})

	t.Run("EmptyEmail", func(t *testing.T) {
		email := ""

		hash, err := generateActivationHash(email)
		require.NoError(t, err)
		require.NotEmpty(t, hash)
	})

	t.Run("ErrorOnRandomBytes", func(t *testing.T) {
		originalReader := RandomBytesReaderMock()
		defer func() { randomBytesReader = originalReader }()

		hash, err := generateActivationHash(email)
		require.Error(t, err)
		require.EqualError(t, err, "could not generate random bytes: mock error")
		require.Empty(t, hash)
	})
}

func TestUsersService_HashPassword(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		service := &UsersService{}
		password := "password123"

		hashedPassword, err := service.HashPassword(password)
		require.NoError(t, err)
		require.NotEmpty(t, hashedPassword)
	})

	t.Run("Error", func(t *testing.T) {
		originalFunc := BcryptGenerateFromPasswordMock()
		defer func() { bcryptGenerateFromPassword = originalFunc }()

		bcryptGenerateFromPassword = func(password []byte, cost int) ([]byte, error) {
			return nil, errors.New("mock bcrypt error")
		}

		service := &UsersService{}
		password := "password123"
		hashedPassword, err := service.HashPassword(password)

		require.Error(t, err)
		require.EqualError(t, err, "mock bcrypt error")
		require.Empty(t, hashedPassword)
	})
}

func TestUsersService_MakeSession(t *testing.T) {
	usersRepo := new(MockUsersRepo)
	sessionsRepo := new(MockSessionsRepo)
	service := NewUsersService(usersRepo, sessionsRepo, nil, "https://example.com")

	t.Run("Success", func(t *testing.T) {
		sessionsRepo.On("Create", mock.AnythingOfType("string"), mock.AnythingOfType("*users.Session")).Return(nil).Once()

		session, err := service.makeSession(1)
		require.NoError(t, err)
		require.NotNil(t, session)
		require.Equal(t, 1, session.UserID)

		sessionsRepo.AssertExpectations(t)
	})

	t.Run("CreateError", func(t *testing.T) {
		sessionsRepo.On("Create", mock.AnythingOfType("string"), mock.AnythingOfType("*users.Session")).Return(fmt.Errorf("mock create session error")).Once()

		session, err := service.makeSession(1)
		require.Error(t, err)
		require.Nil(t, session)
		require.EqualError(t, err, "could not create session: mock create session error")

		sessionsRepo.AssertExpectations(t)
	})
}
