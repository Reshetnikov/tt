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
	usersRepo UsersRepository
	sessionsRepo SessionsRepository
}

// Конструктор для UserService
func NewUsersService(usersRepo UsersRepository, sessionsRepo SessionsRepository) *UsersService {
	return &UsersService{
		usersRepo: usersRepo,
		sessionsRepo: sessionsRepo,
	}
}

type RegisterUserData struct {
	Name                 string
	Email                string
	Password             string
}

var ErrEmailExists = errors.New("email is already in use")

func (s *UsersService) RegisterUser(registerUserData RegisterUserData) (error) {
	existingUser, err := s.usersRepo.GetByEmail(registerUserData.Email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return  ErrEmailExists
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
		Name: registerUserData.Name,
		Email: registerUserData.Email,
		Password: hashedPassword,
		DateAdd: date,
		IsActive: false,
		ActivationHash: activationHash,
		ActivationHashDate: date,
	}
	fmt.Printf("-----USER:%+v\n", user)
	err =  s.usersRepo.Create(user)

	return err
}

func (s *UsersService) ActivateUser(activationHash string) (*Session, error) {
    user, err := s.usersRepo.GetByActivationHash(activationHash)
    if err != nil || user == nil {
        return nil, fmt.Errorf("user not found or activation hash is invalid")
    }
    
    user.IsActive = true
    user.ActivationHash = ""

    err = s.usersRepo.Update(user)
    if err != nil {
        return nil, fmt.Errorf("could not activate user: %w", err)
    }

    session := s.makeSession(user.ID)
	return session, nil
}

// Логика входа
func (s *UsersService) LoginUser(email, password string) (*Session, error) {
	user, err := s.usersRepo.GetByEmail(email)
    if err != nil || user == nil || !checkPasswordHash(password, user.Password) {
        return nil, fmt.Errorf("invalid email or password")
    }
	if (!user.IsActive) {
		return nil, fmt.Errorf("аccount not activated")
	}

	session := s.makeSession(user.ID)
	return session, nil
}

func (s *UsersService) makeSession(userId int) (*Session){
	sessionID := uuid.New().String()
	session := &Session{
		UserID: userId,
		Expiry: time.Now().Add(time.Hour * 24),
	}
	session.SessionID = sessionID
	fmt.Printf("-----SESSION:%+v\n", session)
	s.sessionsRepo.Create(sessionID, session)
    return session
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