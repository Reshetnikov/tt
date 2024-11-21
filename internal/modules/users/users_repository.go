package users

import "time"

type User struct {
	ID                 int       `json:"id" db:"id"`
	Name               string    `json:"name" db:"name"`
	Email              string    `json:"email" db:"email"`
	Password           string    `json:"-" db:"password"`
	TimeZone           string    `json:"timezone" db:"timezone"`
	IsWeekStartMonday  bool      `json:"is_week_start_monday" db:"is_week_start_monday"`
	IsActive           bool      `json:"is_active" db:"is_active"`
	DateAdd            time.Time `json:"date_add" db:"date_add"`
	ActivationHash     string    `json:"activation_hash" db:"activation_hash"`
	ActivationHashDate time.Time `json:"activation_hash_date" db:"activation_hash_date"`
}

type UsersRepository interface {
	Create(user *User) error
	GetByID(id int) *User
	GetByEmail(email string) *User
	GetByActivationHash(activationHash string) *User
	Update(user *User) error
	Delete(id int) error
}
