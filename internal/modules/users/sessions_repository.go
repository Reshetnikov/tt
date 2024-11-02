package users

import (
	"time"
)

type Session struct {
    SessionID string
    UserID int
    Expiry time.Time
}

type SessionsRepository interface {
    Create(sessionID string, session *Session) error
    Get(sessionID string) (*Session, error)
    Delete(sessionID string) error
}