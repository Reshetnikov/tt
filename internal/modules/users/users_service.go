package users

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
	"time-tracker/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

type UsersService struct {
	usersRepo UsersRepository
}

// Конструктор для UserService
func NewUsersService(usersRepo UsersRepository) *UsersService {
	return &UsersService{usersRepo: usersRepo}
}

type RegisterUserData struct {
	Name                 string
	Email                string
	Password             string
}

func (s *UsersService) RegisterUser(registerUserData RegisterUserData) (error) {
	// Логика регистрации пользователя, например:
	// - Валидация данных
	// - Хеширование пароля
	// - Сохранение пользователя в базу данных через userRepo
	existingUser, err := s.usersRepo.GetByEmail(registerUserData.Email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		formErrors := utils.FormErrors{
			"Email": {"Email is already in use"},
		}
		return  formErrors
	}

	hashedPassword, err := hashPassword(registerUserData.Password)
	if err != nil {
		return err
	}
	activationHash, err := generateActivationHash(registerUserData.Email)
	if err != nil {
		return err
	}
	user := &User{
		Name: registerUserData.Name,
		Email: registerUserData.Email,
		Password: hashedPassword,
		DateAdd: time.Now().UTC(),
		IsActive: false,
		ActivationHash: activationHash,
	}
	err =  s.usersRepo.Create(user)

	return err
}

// Логика входа
func (s *UsersService) LoginUser(username, password string) (string, error) {
	// Логика входа: проверка пароля и возврат токена
	return "jwt-token", nil
}

func hashPassword(password string) (string, error) {
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(hashedBytes), nil
}

// Сравнивает хеш пароля с введённым паролем
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