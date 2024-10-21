package repository

import (
	"database/sql"
	"time-tracker/internal/models"
)

type UserRepositoryPostgreSQL struct {
	db *sql.DB
}

func NewUserRepositoryPostgresql(db *sql.DB) *UserRepositoryPostgreSQL {
	return &UserRepositoryPostgreSQL{db: db}
}

func (r *UserRepositoryPostgreSQL) Create(user *models.User) error {
	// Реализация создания пользователя
	return nil
}

func (r *UserRepositoryPostgreSQL) GetByID(id int) (*models.User, error) {
	// Реализация получения пользователя по ID
	return nil, nil
}

func (r *UserRepositoryPostgreSQL) GetByUsername(username string) (*models.User, error) {
	// Реализация получения пользователя по имени
	return nil, nil
}

func (r *UserRepositoryPostgreSQL) Update(user *models.User) error {
	// Реализация обновления пользователя
	return nil
}

func (r *UserRepositoryPostgreSQL) Delete(id int) error {
	// Реализация удаления пользователя
	return nil
}
