package users

import (
	"sync"
	"time"
)

type SessionsRepositoryMem struct {
    mu       sync.Mutex
    sessions map[string]*Session
}

func NewSessionsRepositoryMem() *SessionsRepositoryMem {
    return &SessionsRepositoryMem{
        sessions: make(map[string]*Session),
    }
}

func (repo *SessionsRepositoryMem) Create(sessionID string, session *Session) error {
    repo.mu.Lock()
    defer repo.mu.Unlock()
    repo.sessions[sessionID] = session
    return nil
}

func (repo *SessionsRepositoryMem) Get(sessionID string) (*Session, error) {
    repo.mu.Lock()
    defer repo.mu.Unlock()
    session, exists := repo.sessions[sessionID]
    if !exists || session.Expiry.Before(time.Now()) {
        return nil, nil // Если сессия не существует или истекла
    }
    return session, nil
}

func (repo *SessionsRepositoryMem) Delete(sessionID string) error {
    repo.mu.Lock()
    defer repo.mu.Unlock()
    delete(repo.sessions, sessionID)
    return nil
}