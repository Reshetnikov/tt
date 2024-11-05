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

func (r *UsersRepositoryPostgres) getByField(fieldName string, FieldValue interface{}) (*User, error) {
	validFields := map[string]bool{
		"id":              true,
		"email":           true,
		"activation_hash": true,
	}
	if !validFields[fieldName] {
		return nil, fmt.Errorf("unsupported field: %s", fieldName)
	}
	query := "SELECT id, name, password, email, date_add, activation_hash, activation_hash_date, is_active FROM users WHERE " + fieldName + " = $1"
	rows, _ := r.db.Query(context.Background(), query, FieldValue)
	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by %s = %v: %w", fieldName, FieldValue, err)
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
	fields, placeholders, params := utils.BuildFieldsFromArr(utils.Arr{
		{"name", user.Name},
		{"password", user.Password},
		{"email", user.Email},
		{"date_add", user.DateAdd},
		{"activation_hash", user.ActivationHash},
		{"activation_hash_date", user.ActivationHashDate},
		{"is_active", user.IsActive},
	})
	query := "INSERT INTO users (" + fields + ") VALUES (" + placeholders + ")"
	_, err := r.db.Exec(context.Background(), query, params...)
	if err != nil {
		log.Printf("Failed to insert user: %v", err)
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

func (r *UsersRepositoryPostgres) Update(user *User) error {
	builder := utils.NewBuilderFieldsValues()
	set := builder.BuildFromArr(utils.Arr{
		{"name", user.Name},
		{"password", user.Password},
		{"email", user.Email},
		{"date_add", user.DateAdd},
		{"activation_hash", user.ActivationHash},
		{"activation_hash_date", user.ActivationHashDate},
		{"is_active", user.IsActive},
	})
	where := builder.BuildFromArr(utils.Arr{{"id", user.ID}})
	query := "UPDATE users SET " + set + " WHERE " + where
	_, err := r.db.Exec(context.Background(), query, builder.Params()...)
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
