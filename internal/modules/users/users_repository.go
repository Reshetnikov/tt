package users

import "time"

type User struct {
	ID                 int       `json:"id" db:"id"`
    Name               string    `json:"name" db:"name"`
	Password		   string	 `json:"-"`
    Email              string    `json:"email" db:"email"`
    DateAdd            time.Time `json:"date_add" db:"date_add"`
    ActivationHash     string    `json:"activation_hash" db:"activation_hash"`
    ActivationHashDate time.Time `json:"activation_hash_date" db:"activation_hash_date"`
    IsActive           bool      `json:"is_active" db:"is_active"`
}

type UsersRepository interface {
	Create(user *User) (error)
	GetByID(id int) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByActivationHash(activationHash string) (*User, error)
	Update(user *User) error
	Delete(id int) error
}
