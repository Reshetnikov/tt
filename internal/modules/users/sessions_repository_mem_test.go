//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/users --tags=unit -cover -run TestSessionsRepositoryMem
package users

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSessionsRepositoryMem(t *testing.T) {
	repo := NewSessionsRepositoryMem()

	t.Run("create and get session", func(t *testing.T) {
		sessionID := "session123"
		session := &Session{
			SessionID: sessionID,
			UserID:    1,
			Expiry:    time.Now().Add(time.Hour),
		}

		err := repo.Create(sessionID, session)
		assert.NoError(t, err)

		got, err := repo.Get(sessionID)
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, session, got)
	})

	t.Run("get expired session", func(t *testing.T) {
		sessionID := "expiredSession"
		session := &Session{
			SessionID: sessionID,
			UserID:    2,
			Expiry:    time.Now().Add(-time.Hour), // Expired session
		}

		err := repo.Create(sessionID, session)
		assert.NoError(t, err)

		// Check that the expired session is not available
		got, err := repo.Get(sessionID)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("delete session", func(t *testing.T) {
		sessionID := "sessionToDelete"
		session := &Session{
			SessionID: sessionID,
			UserID:    3,
			Expiry:    time.Now().Add(time.Hour),
		}

		err := repo.Create(sessionID, session)
		assert.NoError(t, err)

		err = repo.Delete(sessionID)
		assert.NoError(t, err)

		got, err := repo.Get(sessionID)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("delete non-existent session", func(t *testing.T) {
		sessionID := "nonExistent"
		err := repo.Delete(sessionID)
		assert.NoError(t, err)
	})
}
