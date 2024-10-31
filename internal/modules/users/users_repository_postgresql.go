package users

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersRepositoryPostgres struct {
	DB *pgxpool.Pool
}

// NewUsersRepositoryPostgres создаёт новый репозиторий пользователей с подключением к базе данных
func NewUsersRepositoryPostgres(db *pgxpool.Pool) *UsersRepositoryPostgres {
	return &UsersRepositoryPostgres{DB: db}
}


func (r *UsersRepositoryPostgres) Create(user *User) error {
	query := `INSERT INTO users (name, password, email, date_add, activation_hash, activation_hash_date, is_active)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.DB.Exec(context.Background(), query, user.Name, user.Password, user.Email, user.DateAdd, user.ActivationHash, user.ActivationHashDate, user.IsActive)
	if err != nil {
		log.Printf("Failed to insert user: %v", err)
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

// GetByID получает пользователя по ID
func (r *UsersRepositoryPostgres) GetByID(id int) (*User, error) {
	query := `SELECT id, name, email, date_add, activation_hash, activation_hash_date, is_active
	FROM users WHERE id = $1`

	var user User
	err := r.DB.QueryRow(context.Background(), query, id).Scan(&user.ID, &user.Name, &user.Email, &user.DateAdd, &user.ActivationHash, &user.ActivationHashDate, &user.IsActive)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		log.Printf("Failed to get user by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to get user by ID %d: %w", id, err)
	}
	return &user, nil
}

// GetByEmail получает пользователя по email
func (r *UsersRepositoryPostgres) GetByEmail(email string) (*User, error) {
	query := `SELECT id, name, password, date_add, activation_hash, activation_hash_date, is_active
	FROM users WHERE email = $1`

	var user User
	err := r.DB.QueryRow(context.Background(), query, email).Scan(&user.ID, &user.Name, &user.Password, &user.DateAdd, &user.ActivationHash, &user.ActivationHashDate, &user.IsActive)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		log.Printf("Failed to get user by email %s: %v", email, err)
		return nil, fmt.Errorf("failed to get user by email %s: %w", email, err)
	}
	return &user, nil
}

// GetByActivationHash получает пользователя по активационному хешу
func (r *UsersRepositoryPostgres) GetByActivationHash(activationHash string) (*User, error) {
	query := `SELECT id, name, email, date_add, is_active
	FROM users WHERE activation_hash = $1`

	var user User
	err := r.DB.QueryRow(context.Background(), query, activationHash).Scan(&user.ID, &user.Name, &user.Email, &user.DateAdd, &user.IsActive)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		log.Printf("Failed to get user by activation hash %s: %v", activationHash, err)
		return nil, fmt.Errorf("failed to get user by activation hash %s: %w", activationHash, err)
	}
	return &user, nil
}

// Update обновляет информацию о пользователе
func (r *UsersRepositoryPostgres) Update(user *User) error {
	query := `UPDATE users SET name = $1, email = $2, password = $3, is_active = $4
	WHERE id = $5`

	_, err := r.DB.Exec(context.Background(), query, user.Name, user.Email, user.Password, user.IsActive, user.ID)
	if err != nil {
		log.Printf("Failed to update user %d: %v", user.ID, err)
		return fmt.Errorf("failed to update user %d: %w", user.ID, err)
	}
	return nil
}

// Delete удаляет пользователя из базы данных
func (r *UsersRepositoryPostgres) Delete(id int) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := r.DB.Exec(context.Background(), query, id)
	if err != nil {
		log.Printf("Failed to delete user %d: %v", id, err)
		return fmt.Errorf("failed to delete user %d: %w", id, err)
	}
	return nil
}