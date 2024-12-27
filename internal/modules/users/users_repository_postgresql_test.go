//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/users --tags=unit -cover -run TestUsersRepositoryPostgres.*
package users

import (
	"fmt"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestUsersRepositoryPostgres_GetByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUsersRepositoryPostgres(mock)
	mock.ExpectQuery(`SELECT id, name, password, timezone, is_week_start_monday, email, date_add, activation_hash, activation_hash_date, is_active FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "password", "timezone", "is_week_start_monday", "email", "date_add", "activation_hash", "activation_hash_date", "is_active"}).
			AddRow(1, "John Doe", "password123", "UTC", true, "john@example.com", time.Now(), "hash123", time.Now(), true))
	user := repo.GetByID(1)
	require.NotNil(t, user)
	require.Equal(t, 1, user.ID)
	require.Equal(t, "John Doe", user.Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersRepositoryPostgres_GetByEmail(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUsersRepositoryPostgres(mock)
	mock.ExpectQuery(`SELECT id, name, password, timezone, is_week_start_monday, email, date_add, activation_hash, activation_hash_date, is_active FROM users WHERE email = \$1`).
		WithArgs("test@example.com").
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "password", "timezone", "is_week_start_monday", "email", "date_add", "activation_hash", "activation_hash_date", "is_active"}).
			AddRow(1, "John Doe", "password123", "UTC", true, "john@example.com", time.Now(), "hash123", time.Now(), true))
	user := repo.GetByEmail("test@example.com")
	require.NotNil(t, user)
	require.Equal(t, 1, user.ID)
	require.Equal(t, "john@example.com", user.Email)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersRepositoryPostgres_GetByActivationHash(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUsersRepositoryPostgres(mock)
	mock.ExpectQuery(`SELECT id, name, password, timezone, is_week_start_monday, email, date_add, activation_hash, activation_hash_date, is_active FROM users WHERE activation_hash = \$1`).
		WithArgs("hash123").
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "password", "timezone", "is_week_start_monday", "email", "date_add", "activation_hash", "activation_hash_date", "is_active"}).
			AddRow(1, "John Doe", "password123", "UTC", true, "john@example.com", time.Now(), "hash123", time.Now(), true))
	user := repo.GetByActivationHash("hash123")
	require.NotNil(t, user)
	require.Equal(t, 1, user.ID)
	require.Equal(t, "hash123", user.ActivationHash)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersRepositoryPostgres_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUsersRepositoryPostgres(mock)
	mock.ExpectExec(`INSERT INTO users`).
		WithArgs("John Doe", "hashed_password", "test@example.com", pgxmock.AnyArg(), "hash123", pgxmock.AnyArg(), false, "UTC", true).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	err = repo.Create(&User{
		Name:               "John Doe",
		Password:           "hashed_password",
		Email:              "test@example.com",
		DateAdd:            time.Now(),
		ActivationHash:     "hash123",
		ActivationHashDate: time.Now(),
		IsActive:           false,
		TimeZone:           "UTC",
		IsWeekStartMonday:  true,
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersRepositoryPostgres_Create_Error(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUsersRepositoryPostgres(mock)

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs("John Doe", "hashed_password", "test@example.com", pgxmock.AnyArg(), "hash123", pgxmock.AnyArg(), false, "UTC", true).
		WillReturnError(fmt.Errorf("database insert error"))

	err = repo.Create(&User{
		Name:               "John Doe",
		Password:           "hashed_password",
		Email:              "test@example.com",
		DateAdd:            time.Now(),
		ActivationHash:     "hash123",
		ActivationHashDate: time.Now(),
		IsActive:           false,
		TimeZone:           "UTC",
		IsWeekStartMonday:  true,
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to insert user")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersRepositoryPostgres_Update(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUsersRepositoryPostgres(mock)
	mock.ExpectExec(`UPDATE users SET`).
		WithArgs("Jane Doe", "new_password", "jane@example.com", pgxmock.AnyArg(), "new_hash", pgxmock.AnyArg(), true, "PST", false, 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	err = repo.Update(&User{
		ID:                 1,
		Name:               "Jane Doe",
		Password:           "new_password",
		Email:              "jane@example.com",
		DateAdd:            time.Now(),
		ActivationHash:     "new_hash",
		ActivationHashDate: time.Now(),
		IsActive:           true,
		TimeZone:           "PST",
		IsWeekStartMonday:  false,
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersRepositoryPostgres_Update_Error(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUsersRepositoryPostgres(mock)
	mock.ExpectExec(`UPDATE users SET`).
		WithArgs("Jane Doe", "new_password", "jane@example.com", pgxmock.AnyArg(), "new_hash", pgxmock.AnyArg(), true, "PST", false, 1).
		WillReturnError(fmt.Errorf("database update error"))
	err = repo.Update(&User{
		ID:                 1,
		Name:               "Jane Doe",
		Password:           "new_password",
		Email:              "jane@example.com",
		DateAdd:            time.Now(),
		ActivationHash:     "new_hash",
		ActivationHashDate: time.Now(),
		IsActive:           true,
		TimeZone:           "PST",
		IsWeekStartMonday:  false,
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to update user")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersRepositoryPostgres_Delete(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUsersRepositoryPostgres(mock)
	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	err = repo.Delete(1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersRepositoryPostgres_Delete_Error(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUsersRepositoryPostgres(mock)
	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnError(fmt.Errorf("database deletee error"))
	err = repo.Delete(1)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to delete user")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersRepositoryPostgres_GetByField_NoRows(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery(`SELECT id, name, password, timezone, is_week_start_monday, email, date_add, activation_hash, activation_hash_date, is_active FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(pgxmock.NewRows([]string{"id"}))
	repo := NewUsersRepositoryPostgres(mock)
	user := repo.GetByID(1)
	require.Nil(t, user)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersRepositoryPostgres_GetByField_ErrorCollectingRow(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery(`SELECT id, name, password, timezone, is_week_start_monday, email, date_add, activation_hash, activation_hash_date, is_active FROM users WHERE id = \$1`).
		WithArgs(1).
		// Some fields were transferred, which causes an error in CollectOneRow
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "email"}).AddRow(1, "Test User", "test@example.com"))

	repo := NewUsersRepositoryPostgres(mock)
	require.Panics(t, func() {
		repo.GetByID(1)
	})
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersRepositoryPostgres_getByField(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUsersRepositoryPostgres(mock)

	t.Run("Valid field and value", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, name, password, timezone, is_week_start_monday, email, date_add, activation_hash, activation_hash_date, is_active FROM users WHERE email = \$1`).
			WithArgs("test@example.com").
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "password", "timezone", "is_week_start_monday", "email", "date_add", "activation_hash", "activation_hash_date", "is_active"}).
				AddRow(1, "John Doe", "password123", "UTC", true, "test@example.com", time.Now(), "hash123", time.Now(), true))

		user := repo.getByField("email", "test@example.com")
		require.NotNil(t, user)
		require.Equal(t, "John Doe", user.Name)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Invalid field name", func(t *testing.T) {
		user := repo.getByField("invalid_field", "value")
		require.Nil(t, user)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("No rows found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, name, password, timezone, is_week_start_monday, email, date_add, activation_hash, activation_hash_date, is_active FROM users WHERE email = \$1`).
			WithArgs("nonexistent@example.com").
			WillReturnRows(pgxmock.NewRows([]string{"id"}))

		user := repo.getByField("email", "nonexistent@example.com")
		require.Nil(t, user)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Query execution error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, name, password, timezone, is_week_start_monday, email, date_add, activation_hash, activation_hash_date, is_active FROM users WHERE email = \$1`).
			WithArgs("error@example.com").
			WillReturnError(fmt.Errorf("query failed"))

		user := repo.getByField("email", "error@example.com")
		require.Nil(t, user)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
