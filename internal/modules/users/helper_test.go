// For all go:build
// If a function is defined in a file without a build tag, but is used in a file with a build tag, it is considered unused. Therefore, functions defined here are public.
package users

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/stretchr/testify/mock"
)

func SetAppDir() {
	os.Chdir("/app")
}

func BadRequestPost(url string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader("%%%"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

// Mock randomBytesReader and return original.
func RandomBytesReaderMock() func(_ []byte) (int, error) {
	originalReader := randomBytesReader
	randomBytesReader = func(_ []byte) (int, error) {
		return 0, fmt.Errorf("mock error")
	}
	return originalReader
}

// Mock bcryptGenerateFromPassword and return original.
func BcryptGenerateFromPasswordMock() func(password []byte, cost int) ([]byte, error) {
	originalFunc := bcryptGenerateFromPassword
	bcryptGenerateFromPassword = func(password []byte, cost int) ([]byte, error) {
		return nil, fmt.Errorf("mock bcrypt error")
	}
	return originalFunc
}

type MockUsersService struct {
	mock.Mock
}

func (m *MockUsersService) ActivateUser(activationHash string) (*Session, error) {
	args := m.Called(activationHash)
	session, _ := args.Get(0).(*Session)
	return session, args.Error(1)
}
func (m *MockUsersService) RegisterUser(registerUserData RegisterUserData) error {
	args := m.Called(registerUserData)
	return args.Error(0)
}
func (m *MockUsersService) LoginWithToken(token string) (*Session, error) {
	args := m.Called(token)
	session, _ := args.Get(0).(*Session)
	return session, args.Error(1)
}
func (m *MockUsersService) LoginUser(email, password string) (*Session, error) {
	args := m.Called(email, password)
	session, _ := args.Get(0).(*Session)
	return session, args.Error(1)
}
func (m *MockUsersService) LogoutUser(sessionID string) error {
	args := m.Called(sessionID)
	return args.Error(0)
}
func (m *MockUsersService) SendLinkToLogin(email string) (int, error) {
	args := m.Called(email)
	return args.Int(0), args.Error(1)
}
func (m *MockUsersService) ReSendActivationEmail(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}
func (m *MockUsersService) UserGetByEmail(email string) *User {
	args := m.Called(email)
	if user, ok := args.Get(0).(*User); ok {
		return user
	}
	return nil
}
func (m *MockUsersService) UserUpdate(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}
func (m *MockUsersService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

type MockUsersRepo struct {
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

type MockSessionsRepo struct {
	mock.Mock
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

type MockMailService struct {
	mock.Mock
}

func (m *MockMailService) SendActivationEmail(email, name, link string) error {
	args := m.Called(email, name, link)
	return args.Error(0)
}

func (m *MockMailService) SendLoginWithTokenEmail(email, name, link string) error {
	args := m.Called(email, name, link)
	return args.Error(0)
}
