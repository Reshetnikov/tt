package users

type UserService struct {
	userRepo UserRepository
}

// Конструктор для UserService
func NewUserService(userRepo UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// Логика регистрации
func (s *UserService) RegisterUser(username, password string) error {
	// Логика регистрации пользователя, например:
	// - Валидация данных
	// - Хеширование пароля
	// - Сохранение пользователя в базу данных через userRepo
	return nil
}

// Логика входа
func (s *UserService) LoginUser(username, password string) (string, error) {
	// Логика входа: проверка пароля и возврат токена
	return "jwt-token", nil
}
