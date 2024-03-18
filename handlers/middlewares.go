package handlers

import (
	"FilmCollection/db"
	"context"
	"log/slog"
	"net/http"
)

func Wrap(f http.HandlerFunc) http.HandlerFunc {
	for _, mw := range []func(http.HandlerFunc) http.HandlerFunc{
		authMiddleware,
	} {
		f = mw(f)
	}

	return f
}

func authMiddleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "authorization error", http.StatusUnauthorized)
			slog.Error("Authorization error: ", "error", "no basic auth", "status", http.StatusUnauthorized)
			return
		}

		var id int
		var admin bool
		err := db.Conn.QueryRow("SELECT id, admin FROM Users WHERE login = $1 AND password = $2", user, pass).Scan(&id, &admin)
		if err != nil || id == 0 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "authorization error", http.StatusUnauthorized)
			slog.Error("Authorization error: ", "error", err, "status", http.StatusUnauthorized)
			return
		}

		// Add user to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user", id)
		ctx = context.WithValue(ctx, "admin", admin)
		r = r.WithContext(ctx)

		f(w, r)
	}
}
