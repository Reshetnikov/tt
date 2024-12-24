package users

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionsRepositoryRedis struct {
	client      *redis.Client
	jsonMarshal func(v any) ([]byte, error)
}

func NewSessionsRepositoryRedis(client *redis.Client) *SessionsRepositoryRedis {
	return &SessionsRepositoryRedis{
		client:      client,
		jsonMarshal: json.Marshal,
	}
}

func (repo *SessionsRepositoryRedis) Create(sessionID string, session *Session) error {
	ctx := context.Background()
	data, err := repo.jsonMarshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	expiration := time.Until(session.Expiry)
	err = repo.client.Set(ctx, sessionID, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set session in Redis: %w", err)
	}
	return nil
}

func (repo *SessionsRepositoryRedis) Get(sessionID string) (*Session, error) {
	ctx := context.Background()
	data, err := repo.client.Get(ctx, sessionID).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get session from Redis: %w", err)
	}

	var session Session
	err = json.Unmarshal([]byte(data), &session)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	if session.Expiry.Before(time.Now()) {
		return nil, nil
	}

	return &session, nil
}

func (repo *SessionsRepositoryRedis) Delete(sessionID string) error {
	ctx := context.Background()
	err := repo.client.Del(ctx, sessionID).Err()
	if err != nil {
		return fmt.Errorf("failed to delete session from Redis: %w", err)
	}
	return nil
}
