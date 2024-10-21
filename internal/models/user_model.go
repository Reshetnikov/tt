package models

// User представляет модель пользователя
type User struct {
	ID       int    `json:"id"`       // Идентификатор пользователя
	Username string `json:"username"` // Имя пользователя
	Password string `json:"password"` // Пароль пользователя (возможно, хешированный)
	Email    string `json:"email"`    // Электронная почта пользователя
}
