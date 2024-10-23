package users

import (
	"net/http"
	"time-tracker/internal/utils"
)

type UserHandler struct {
	userService *UserService // Зависимость от сервиса пользователя
}

// Конструктор для UserHandler
func NewUserHandler(userService *UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Обработчик для регистрации
func (h *UserHandler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "signup", map[string]interface{}{
		"Title": "Sign Up",
	})
}

// Обработчик для входа
func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Логика входа пользователя с использованием userService
}
