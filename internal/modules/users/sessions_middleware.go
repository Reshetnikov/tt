package users

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type ContextKey string

const ContextUserKey ContextKey = "user"

func SessionMiddleware(next http.Handler, sessionsRepo SessionsRepository, usersRepo UsersRepository) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("session_id")
        slog.Debug("SessionMiddleware", "cookie", cookie)
        if err != nil || cookie.Value == "" {
            next.ServeHTTP(w, r)
            return
        }

        session, err := sessionsRepo.Get(cookie.Value)
        slog.Debug("SessionMiddleware", "session", session)
        if err != nil || session == nil || session.Expiry.Before(time.Now()) {
            next.ServeHTTP(w, r)
            return
        }

        user, err := usersRepo.GetByID(session.UserID)
        slog.Debug("SessionMiddleware", "user", user)
        if err == nil && user != nil {
            ctx := context.WithValue(r.Context(), ContextUserKey, user)
            r = r.WithContext(ctx)
        }

        next.ServeHTTP(w, r)
    })
}

func GetUserFromRequest(r *http.Request) (*User) {
    userAny := r.Context().Value(ContextUserKey)
    if userAny == nil {
        return nil
    }
    user, err := userAny.(User)
    if err {
        return nil
    }
    return &user
}