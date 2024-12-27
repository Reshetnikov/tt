//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/users --tags=unit -cover -run TestUsersRepositoryMem.*
package users

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUsersRepositoryMem_Create(t *testing.T) {
	repo := NewUsersRepositoryMem()

	user := &User{Email: "test@example.com"}
	err := repo.Create(user)
	require.NoError(t, err)
	require.Equal(t, 1, user.ID)
}

func TestUsersRepositoryMem_GetByID(t *testing.T) {
	repo := NewUsersRepositoryMem()

	user := &User{Email: "test@example.com"}
	_ = repo.Create(user)

	found := repo.GetByID(user.ID)
	require.NotNil(t, found)
	require.Equal(t, user.Email, found.Email)

	notFound := repo.GetByID(999)
	require.Nil(t, notFound)
}

func TestUsersRepositoryMem_GetByEmail(t *testing.T) {
	repo := NewUsersRepositoryMem()

	user := &User{Email: "test@example.com"}
	_ = repo.Create(user)

	found := repo.GetByEmail(user.Email)
	require.NotNil(t, found)
	require.Equal(t, user.ID, found.ID)

	notFound := repo.GetByEmail("nonexistent@example.com")
	require.Nil(t, notFound)
}

func TestUsersRepositoryMem_GetByActivationHash(t *testing.T) {
	repo := NewUsersRepositoryMem()

	user := &User{ActivationHash: "hash123"}
	_ = repo.Create(user)

	found := repo.GetByActivationHash(user.ActivationHash)
	require.NotNil(t, found)
	require.Equal(t, user.ID, found.ID)

	notFound := repo.GetByActivationHash("nonexistent")
	require.Nil(t, notFound)
}

func TestUsersRepositoryMem_Update(t *testing.T) {
	repo := NewUsersRepositoryMem()

	user := &User{Email: "test@example.com"}
	_ = repo.Create(user)

	user.Email = "updated@example.com"
	err := repo.Update(user)
	require.NoError(t, err)

	updated := repo.GetByID(user.ID)
	require.Equal(t, "updated@example.com", updated.Email)

	nonexistentUser := &User{ID: 999}
	err = repo.Update(nonexistentUser)
	require.Error(t, err)
}

func TestUsersRepositoryMem_Delete(t *testing.T) {
	repo := NewUsersRepositoryMem()

	user := &User{Email: "test@example.com"}
	_ = repo.Create(user)

	err := repo.Delete(user.ID)
	require.NoError(t, err)

	deleted := repo.GetByID(user.ID)
	require.Nil(t, deleted)

	err = repo.Delete(999)
	require.Error(t, err)
}
