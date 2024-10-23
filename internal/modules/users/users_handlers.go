package users

import (
	"net/http"
	"time-tracker/internal/utils"
)

type UsersHandler struct {
	userService *UsersService // Зависимость от сервиса пользователя
}

// Конструктор для UserHandler
func NewUserHandlers(userService *UsersService) *UsersHandler {
	return &UsersHandler{userService: userService}
}

// Обработчик для регистрации
func (h *UsersHandler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "signup", map[string]interface{}{
		"Title": "Sign Up",
	})
}

// Обработчик для входа
func (h *UsersHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Логика входа пользователя с использованием userService
}
