package users

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UsersService struct {
	usersRepo    UsersRepository
	sessionsRepo SessionsRepository
}

// Конструктор для UserService
func NewUsersService(usersRepo UsersRepository, sessionsRepo SessionsRepository) *UsersService {
	return &UsersService{
		usersRepo:    usersRepo,
		sessionsRepo: sessionsRepo,
	}
}

type RegisterUserData struct {
	Name              string
	Email             string
	Password          string
	TimeZone          string
	IsWeekStartMonday bool
}

var ErrEmailExists = errors.New("email is already in use")
var ErrAccountNotActivated = errors.New("account not activated")
var ErrInvalidEmailOrPassword = errors.New("invalid email or password")
var ErrUserNotFoundOrActivationHashIsInvalid = errors.New("user not found or activation hash is invalid")
var ErrUserNotFound = errors.New("user not found")
var ErrTimeUntilResend = errors.New("please wait before resending")

func (s *UsersService) RegisterUser(registerUserData RegisterUserData) error {
	existingUser := s.usersRepo.GetByEmail(registerUserData.Email)
	if existingUser != nil {
		if !existingUser.IsActive {
			return ErrAccountNotActivated
		}
		return ErrEmailExists
	}

	hashedPassword, err := hashPassword(registerUserData.Password)
	if err != nil {
		return err
	}
	activationHash, err := generateActivationHash(registerUserData.Email)
	if err != nil {
		return err
	}
	date := time.Now().UTC()
	user := &User{
		Name:               registerUserData.Name,
		Email:              registerUserData.Email,
		Password:           hashedPassword,
		TimeZone:           registerUserData.TimeZone,
		IsWeekStartMonday:  registerUserData.IsWeekStartMonday,
		IsActive:           false,
		DateAdd:            date,
		ActivationHash:     activationHash,
		ActivationHashDate: date,
	}
	err = s.usersRepo.Create(user)
	if err == nil {
		s.sendActivationEmail(user)
	}

	return err
}

func (s *UsersService) ActivateUser(activationHash string) (*Session, error) {
	user := s.usersRepo.GetByActivationHash(activationHash)
	if user == nil {
		return nil, ErrUserNotFoundOrActivationHashIsInvalid
	}

	user.IsActive = true
	user.ActivationHash = ""

	err := s.usersRepo.Update(user)
	if err != nil {
		return nil, fmt.Errorf("could not activate user: %w", err)
	}

	session, err := s.makeSession(user.ID)
	return session, err
}

// Логика входа
func (s *UsersService) LoginUser(email, password string) (*Session, error) {
	user := s.usersRepo.GetByEmail(email)
	if user == nil || !checkPasswordHash(password, user.Password) {
		return nil, ErrInvalidEmailOrPassword
	}
	if !user.IsActive {
		return nil, ErrAccountNotActivated
	}
	session, err := s.makeSession(user.ID)
	return session, err
}

func (s *UsersService) LogoutUser(sessionID string) error {
	return s.sessionsRepo.Delete(sessionID)
}

func (s *UsersService) SendLinkToLogin(email string) (timeUntilResend int, err error) {
	user := s.usersRepo.GetByEmail(email)
	if user == nil {
		return 0, ErrUserNotFound
	}

	if !user.IsActive {
		return 0, ErrAccountNotActivated
	}

	timeUntilResend = user.TimeUntilResend()
	if timeUntilResend > 0 {
		return timeUntilResend, ErrTimeUntilResend
	}

	activationHash, err := generateActivationHash(user.Email)
	if err != nil {
		return 0, err
	}
	user.ActivationHash = activationHash
	user.ActivationHashDate = time.Now().UTC()
	err = s.usersRepo.Update(user)
	if err != nil {
		return 0, err
	}
	s.sendLoginEmail(user)

	return user.TimeUntilResend(), nil
}

func (s *UsersService) ReSendActivationEmail(user *User) error {

	activationHash, err := generateActivationHash(user.Email)
	if err != nil {
		return err
	}
	user.ActivationHash = activationHash
	user.ActivationHashDate = time.Now().UTC()
	err = s.usersRepo.Update(user)
	if err != nil {
		return err
	}
	s.sendActivationEmail(user)
	return nil
}

func (s *UsersService) sendActivationEmail(user *User) {
	D("Sending activation", "user", user)
}
func (s *UsersService) sendLoginEmail(user *User) {
	D("Sending Login Email", "user", user)
}

func (s *UsersService) makeSession(userId int) (*Session, error) {
	sessionID := uuid.New().String()
	session := &Session{
		UserID: userId,
		Expiry: time.Now().AddDate(1, 0, 0),
	}
	session.SessionID = sessionID
	err := s.sessionsRepo.Create(sessionID, session)
	if err != nil {
		return nil, fmt.Errorf("could not create session: %w", err)
	}
	return session, nil
}

func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func checkPasswordHash(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func generateActivationHash(email string) (string, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("could not generate random bytes: %w", err)
	}

	data := append(randomBytes, []byte(email)...)

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}
