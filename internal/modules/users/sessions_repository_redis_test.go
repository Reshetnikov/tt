//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/modules/users --tags=unit -cover -run TestSessionsRepositoryRedis.*
package users

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionsRepositoryRedis_Create(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		repo := NewSessionsRepositoryRedis(client)

		session := &Session{
			SessionID: "session123",
			UserID:    1,
			Expiry:    time.Now().Add(time.Hour).UTC(),
		}

		data, err := json.Marshal(session)
		require.NoError(t, err)

		mock.ExpectSet(
			session.SessionID,
			data,
			time.Until(session.Expiry),
		).SetVal("OK")

		err = repo.Create(session.SessionID, session)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("redis error", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		repo := NewSessionsRepositoryRedis(client)

		session := &Session{
			SessionID: "session123",
			UserID:    1,
			Expiry:    time.Now().Add(time.Hour).UTC(),
		}

		data, err := json.Marshal(session)
		require.NoError(t, err)

		mock.ExpectSet(
			session.SessionID,
			data,
			time.Until(session.Expiry),
		).SetErr(errors.New("redis error"))

		err = repo.Create(session.SessionID, session)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to set session in Redis")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("marshal error", func(t *testing.T) {
		client, _ := redismock.NewClientMock()
		repo := NewSessionsRepositoryRedis(client)

		// Replace json.Marshal with a function that always returns an error
		repo.jsonMarshal = func(v any) ([]byte, error) {
			return nil, errors.New("mock marshal error")
		}

		session := &Session{
			SessionID: "session123",
			UserID:    1,
			Expiry:    time.Now().Add(time.Hour).UTC(),
		}

		err := repo.Create(session.SessionID, session)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal session")
	})
}

func TestSessionsRepositoryRedis_Get(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		repo := NewSessionsRepositoryRedis(client)

		session := &Session{
			SessionID: "session123",
			UserID:    1,
			Expiry:    time.Now().Add(time.Hour).UTC(),
		}

		data, err := json.Marshal(session)
		require.NoError(t, err)

		mock.ExpectGet("session123").SetVal(string(data))

		result, err := repo.Get("session123")
		assert.NoError(t, err)
		assert.Equal(t, session.UserID, result.UserID)
		assert.Equal(t, session.SessionID, result.SessionID)
		assert.Equal(t, session.Expiry.Unix(), result.Expiry.Unix()) // Сравниваем Unix timestamp
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("non-existent session", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		repo := NewSessionsRepositoryRedis(client)

		mock.ExpectGet("session123").SetErr(redis.Nil)

		result, err := repo.Get("session123")
		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("redis error", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		repo := NewSessionsRepositoryRedis(client)

		mock.ExpectGet("session123").SetErr(errors.New("redis error"))

		result, err := repo.Get("session123")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get session from Redis")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("unmarshal error", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		repo := NewSessionsRepositoryRedis(client)

		mock.ExpectGet("session123").SetVal("invalid json")

		result, err := repo.Get("session123")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to unmarshal session")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("expired session", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		repo := NewSessionsRepositoryRedis(client)

		session := &Session{
			SessionID: "session123",
			UserID:    1,
			Expiry:    time.Now().Add(-time.Hour).UTC(),
		}

		data, err := json.Marshal(session)
		require.NoError(t, err)

		mock.ExpectGet("session123").SetVal(string(data))

		result, err := repo.Get("session123")
		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSessionsRepositoryRedis_Delete(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		repo := NewSessionsRepositoryRedis(client)

		mock.ExpectDel("session123").SetVal(1)

		err := repo.Delete("session123")
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("redis error", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		repo := NewSessionsRepositoryRedis(client)

		mock.ExpectDel("session123").SetErr(errors.New("redis error"))

		err := repo.Delete("session123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete session from Redis")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
