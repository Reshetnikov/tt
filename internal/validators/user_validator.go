package validators

import (
	"errors"
	"time-tracker/internal/models"
)

// ValidateUser проверяет, валидны ли данные пользователя
func ValidateUser(u *models.User) error {
	if u.Username == "" {
		return errors.New("username is required")
	}
	if u.Password == "" {
		return errors.New("password is required")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	// Можно добавить дополнительную валидацию, например, на формат email
	return nil
}
