package users

type UsersHandler struct {
	usersService *UsersService 
}

func NewUsersHandlers(usersService *UsersService) *UsersHandler {
	return &UsersHandler{
		usersService: usersService,
	}
}
