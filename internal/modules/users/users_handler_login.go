package users

import "net/http"

// Обработчик для входа
func (h *UsersHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Логика входа пользователя с использованием userService
}