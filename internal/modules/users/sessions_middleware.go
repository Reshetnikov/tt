package users

import (
	"context"
	"net/http"
	"time"
)

type ContextKey string

const ContextUserKey ContextKey = "user"

// SessionMiddleware проверяет сессию в cookie и добавляет пользователя в контекст запроса
func (s *UsersService) SessionMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("session_id")
        if err != nil || cookie.Value == "" {
            next.ServeHTTP(w, r)
            return
        }

        session, err := s.sessionsRepo.Get(cookie.Value)
        if err != nil || session == nil || session.Expiry.Before(time.Now()) {
            next.ServeHTTP(w, r)
            return
        }

        user, err := s.usersRepo.GetByID(session.UserID)
        if err == nil && user != nil {
            ctx := context.WithValue(r.Context(), ContextUserKey, user)
            r = r.WithContext(ctx)
        }

        next.ServeHTTP(w, r)
    })
}