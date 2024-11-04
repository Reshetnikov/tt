package users

import (
	"context"
	"log/slog"
	"net/http"
	"time"
	"time-tracker/internal/middleware"
)

// SessionMiddleware проверяет сессию в cookie и добавляет пользователя в контекст запроса
func (s *UsersService) SessionMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("session_id")
        slog.Debug("SessionMiddleware", "cookie", cookie)
        if err != nil || cookie.Value == "" {
            next.ServeHTTP(w, r)
            return
        }

        session, err := s.sessionsRepo.Get(cookie.Value)
        slog.Debug("SessionMiddleware", "session", session)
        if err != nil || session == nil || session.Expiry.Before(time.Now()) {
            next.ServeHTTP(w, r)
            return
        }

        user, err := s.usersRepo.GetByID(session.UserID)
        slog.Debug("SessionMiddleware", "user", user)
        if err == nil && user != nil {
            ctx := context.WithValue(r.Context(), middleware.ContextUserKey, user)
            r = r.WithContext(ctx)
        }

        next.ServeHTTP(w, r)
    })
}