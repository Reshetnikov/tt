package users

import (
	"context"
	"fmt"
	"log"
	"time-tracker/internal/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersRepositoryPostgres struct {
	db *pgxpool.Pool
}

// NewUsersRepositoryPostgres создаёт новый репозиторий пользователей с подключением к базе данных
func NewUsersRepositoryPostgres(db *pgxpool.Pool) *UsersRepositoryPostgres {
	return &UsersRepositoryPostgres{db: db}
}

func (r *UsersRepositoryPostgres) getByField(field string, value interface{}) (*User, error) {
    validFields := map[string]bool{
        "id":                true,
        "email":             true,
        "activation_hash":   true,
    }
    if !validFields[field] {
        return nil, fmt.Errorf("unsupported field: %s", field)
    }
    query := fmt.Sprintf(`SELECT id, name, password, email, date_add, activation_hash, activation_hash_date, is_active FROM users WHERE %s = $1`, field)
	rows, _ := r.db.Query(context.Background(), query, value)
	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	utils.Dump("user", user)
	// query := fmt.Sprintf(`SELECT id, name, password, email, date_add, activation_hash, activation_hash_date, is_active 
    //                       FROM users WHERE %s = $1`, field)
    // user := &User{}
    // err := r.db.QueryRow(context.Background(), query, value).Scan(
    //     &user.ID,
    //     &user.Name,
	// 	&user.Password,
    //     &user.Email,
    //     &user.DateAdd,
    //     &user.ActivationHash,
    //     &user.ActivationHashDate,
    //     &user.IsActive,
    // )

    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to get user by %s = %v: %w", field, value, err)
    }
    return &user, nil
}

func (r *UsersRepositoryPostgres) GetByID(id int) (*User, error) {
	return r.getByField("id", id)
}

func (r *UsersRepositoryPostgres) GetByEmail(email string) (*User, error) {
	return r.getByField("email", email)
}

func (r *UsersRepositoryPostgres) GetByActivationHash(activationHash string) (*User, error) {
	return r.getByField("activation_hash", activationHash)
}

func (r *UsersRepositoryPostgres) Create(user *User) error {
	query := `INSERT INTO users (name, password, email, date_add, activation_hash, activation_hash_date, is_active)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.Exec(context.Background(), query, user.Name, user.Password, user.Email, user.DateAdd, user.ActivationHash, user.ActivationHashDate, user.IsActive)
	if err != nil {
		log.Printf("Failed to insert user: %v", err)
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

func (r *UsersRepositoryPostgres) Update(user *User) error {
	query := `UPDATE users SET name = $1, password = $2, email = $3, date_add = $4, activation_hash = $5, activation_hash_date = $6, is_active = $7
	WHERE id = $8`

	_, err := r.db.Exec(context.Background(), query, user.Name, user.Password, user.Email, user.DateAdd, user.ActivationHash, user.ActivationHashDate, user.IsActive, user.ID)
	if err != nil {
		log.Printf("Failed to update user %d: %v", user.ID, err)
		return fmt.Errorf("failed to update user %d: %w", user.ID, err)
	}
	return nil
}

func (r *UsersRepositoryPostgres) Delete(id int) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := r.db.Exec(context.Background(), query, id)
	if err != nil {
		log.Printf("Failed to delete user %d: %v", id, err)
		return fmt.Errorf("failed to delete user %d: %w", id, err)
	}
	return nil
}