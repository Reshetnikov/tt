package users

import "time"

type User struct {
	ID				int			`json:"id"`
	Name			string		`json:"username"`
	Password		string		`json:"-"`
	Email			string		`json:"email"`
	DateAdd			time.Time	`json:"date_add"`
	ActivationHash	string		`json:"-"`
	IsActive		bool		`json:"-"`
}

type UsersRepository interface {
	Create(user *User) (error)
	GetByID(id int) (*User, error)
	GetByEmail(email string) (*User, error)
	Update(user *User) error
	Delete(id int) error
}
