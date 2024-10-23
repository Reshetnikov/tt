package users

import (
	"net/http"
	"time-tracker/internal/utils"
)

type UsersHandler struct {
	usersService *UsersService // Зависимость от сервиса пользователя
}

// Конструктор для UserHandler
func NewUsersHandlers(usersService *UsersService) *UsersHandler {
	return &UsersHandler{usersService: usersService}
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
