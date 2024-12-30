package users

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UsersService struct {
	usersRepo    UsersRepository
	sessionsRepo SessionsRepository
	mailService  MailService
	siteUrl      string
}

func NewUsersService(usersRepo UsersRepository, sessionsRepo SessionsRepository, mailService MailService, siteUrl string) *UsersService {
	return &UsersService{
		usersRepo:    usersRepo,
		sessionsRepo: sessionsRepo,
		mailService:  mailService,
		siteUrl:      siteUrl,
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

var randomBytesReader = rand.Read
var bcryptGenerateFromPassword = bcrypt.GenerateFromPassword

func (s *UsersService) RegisterUser(registerUserData RegisterUserData) error {
	existingUser := s.usersRepo.GetByEmail(registerUserData.Email)
	if existingUser != nil {
		if !existingUser.IsActive {
			return ErrAccountNotActivated
		}
		return ErrEmailExists
	}

	hashedPassword, err := s.HashPassword(registerUserData.Password)
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
	if err != nil {
		return err
	}
	go func() {
		err := s.mailService.SendActivationEmail(user.Email, user.Name, s.activationLink(user.ActivationHash))
		if err != nil {
			slog.Error("Failed to send activation email.", "err", err)
		}
	}()
	return nil
}

func (s *UsersService) ActivateUser(activationHash string) (*Session, error) {
	user := s.usersRepo.GetByActivationHash(activationHash)
	if user == nil || time.Since(user.ActivationHashDate).Minutes() > 15 {
		return nil, ErrUserNotFound
	}

	user.IsActive = true
	user.ActivationHash = ""

	err := s.usersRepo.Update(user)
	if err != nil {
		return nil, fmt.Errorf("could not Update user: %w", err)
	}

	session, err := s.makeSession(user.ID)
	return session, err
}

func (s *UsersService) LoginWithToken(token string) (*Session, error) {
	user := s.usersRepo.GetByActivationHash(token)
	if user == nil || time.Since(user.ActivationHashDate).Minutes() > 15 {
		return nil, ErrUserNotFound
	}

	if !user.IsActive {
		return nil, ErrAccountNotActivated
	}

	user.ActivationHash = ""

	err := s.usersRepo.Update(user)
	if err != nil {
		return nil, fmt.Errorf("could not Update user: %w", err)
	}

	session, err := s.makeSession(user.ID)
	return session, err
}

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
	go func() {
		err := s.mailService.SendLoginWithTokenEmail(user.Email, user.Name, s.loginWithTokenLink(user.ActivationHash))
		if err != nil {
			slog.Error("Failed to send activation email.", "err", err)
		}
	}()

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
	go func() {
		err := s.mailService.SendActivationEmail(user.Email, user.Name, s.activationLink(user.ActivationHash))
		if err != nil {
			slog.Error("Failed to send activation email.", "err", err)
		}
	}()
	return nil
}

func (s *UsersService) UserGetByEmail(email string) *User {
	return s.usersRepo.GetByEmail(email)
}

func (s *UsersService) UserUpdate(user *User) error {
	return s.usersRepo.Update(user)
}

func (s *UsersService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcryptGenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
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

func (s *UsersService) activationLink(hash string) (link string) {
	return fmt.Sprintf("%s/activation?hash=%s", s.siteUrl, hash)
}
func (s *UsersService) loginWithTokenLink(token string) (link string) {
	return fmt.Sprintf("%s/login-with-token?token=%s", s.siteUrl, token)
}

func checkPasswordHash(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func generateActivationHash(email string) (string, error) {
	randomBytes := make([]byte, 16)
	_, err := randomBytesReader(randomBytes)
	if err != nil {
		return "", fmt.Errorf("could not generate random bytes: %w", err)
	}

	data := append(randomBytes, []byte(email)...)

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}
