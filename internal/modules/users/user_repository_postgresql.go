package users

import (
	"database/sql"
)

type UserRepositoryPostgreSQL struct {
	db *sql.DB
}

func NewUserRepositoryPostgresql(db *sql.DB) *UserRepositoryPostgreSQL {
	return &UserRepositoryPostgreSQL{db: db}
}

func (r *UserRepositoryPostgreSQL) Create(user *User) error {
	// Реализация создания пользователя
	return nil
}

func (r *UserRepositoryPostgreSQL) GetByID(id int) (*User, error) {
	// Реализация получения пользователя по ID
	return nil, nil
}

func (r *UserRepositoryPostgreSQL) GetByUsername(username string) (*User, error) {
	// Реализация получения пользователя по имени
	return nil, nil
}

func (r *UserRepositoryPostgreSQL) Update(user *User) error {
	// Реализация обновления пользователя
	return nil
}

func (r *UserRepositoryPostgreSQL) Delete(id int) error {
	// Реализация удаления пользователя
	return nil
}
