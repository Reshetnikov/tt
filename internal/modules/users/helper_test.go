// For all go:build
// If a function is defined in a file without a build tag, but is used in a file with a build tag, it is considered unused. Therefore, functions defined here are public.
package users

import (
	"os"

	"github.com/stretchr/testify/mock"
)

func SetAppDir() {
	os.Chdir("/app")
}

type MockUsersService struct {
	mock.Mock
}

func (m *MockUsersService) ActivateUser(activationHash string) (*Session, error) {
	args := m.Called(activationHash)
	return args.Get(0).(*Session), args.Error(1)
}
func (m *MockUsersService) RegisterUser(registerUserData RegisterUserData) error {
	return nil
}
func (m *MockUsersService) LoginWithToken(token string) (*Session, error) {
	args := m.Called(token)
	return args.Get(0).(*Session), args.Error(1)
}
func (m *MockUsersService) LoginUser(email, password string) (*Session, error) {
	args := m.Called(email, password)
	return args.Get(0).(*Session), args.Error(1)
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
	return nil
}
func (m *MockUsersService) UserGetByEmail(email string) *User {
	return nil
}
func (m *MockUsersService) UserUpdate(user *User) error {
	return nil
}
