package users

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"time-tracker/internal/middleware"
)

// SessionMiddleware проверяет сессию в cookie и добавляет пользователя в контекст запроса
func (s *UsersService) SessionMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("session_id")
        fmt.Printf("-----SessionMiddleware cookie:%+v\n", cookie)
        if err != nil || cookie.Value == "" {
            next.ServeHTTP(w, r)
            return
        }

        session, err := s.sessionsRepo.Get(cookie.Value)
        fmt.Printf("-----SessionMiddleware session:%+v\n", session)
        if err != nil || session == nil || session.Expiry.Before(time.Now()) {
            next.ServeHTTP(w, r)
            return
        }

        user, err := s.usersRepo.GetByID(session.UserID)
        fmt.Printf("-----SessionMiddleware USER:%+v\n", user)
        if err == nil && user != nil {
            ctx := context.WithValue(r.Context(), middleware.ContextUserKey, user)
            r = r.WithContext(ctx)
        }

        next.ServeHTTP(w, r)
    })
}