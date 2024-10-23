package users

type UsersService struct {
	usersRepo UsersRepository
}

// Конструктор для UserService
func NewUserService(usersRepo UsersRepository) *UsersService {
	return &UsersService{usersRepo: usersRepo}
}

// Логика регистрации
func (s *UsersService) RegisterUser(username, password string) error {
	// Логика регистрации пользователя, например:
	// - Валидация данных
	// - Хеширование пароля
	// - Сохранение пользователя в базу данных через userRepo
	return nil
}

// Логика входа
func (s *UsersService) LoginUser(username, password string) (string, error) {
	// Логика входа: проверка пароля и возврат токена
	return "jwt-token", nil
}
