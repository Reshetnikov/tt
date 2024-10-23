package users

import (
	"database/sql"
)

type UsersRepositoryPostgreSQL struct {
	db *sql.DB
}

func NewUsersRepositoryPostgresql(db *sql.DB) *UsersRepositoryPostgreSQL {
	return &UsersRepositoryPostgreSQL{db: db}
}

func (r *UsersRepositoryPostgreSQL) Create(user *User) error {
	// Реализация создания пользователя
	return nil
}

func (r *UsersRepositoryPostgreSQL) GetByID(id int) (*User, error) {
	// Реализация получения пользователя по ID
	return nil, nil
}

func (r *UsersRepositoryPostgreSQL) GetByEmail(email string) (*User, error) {
	// Реализация получения пользователя по имени
	return nil, nil
}

func (r *UsersRepositoryPostgreSQL) Update(user *User) error {
	// Реализация обновления пользователя
	return nil
}

func (r *UsersRepositoryPostgreSQL) Delete(id int) error {
	// Реализация удаления пользователя
	return nil
}
