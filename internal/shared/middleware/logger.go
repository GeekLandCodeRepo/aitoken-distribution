package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"llm-gateway/internal/shared/uuid"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewV7String()
		}
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

func Logger(next http.Handler) http.Handler {
	return middleware.Logger(next)
}

func Recoverer(next http.Handler) http.Handler {
	return middleware.Recoverer(next)
}

func RealIP(next http.Handler) http.Handler {
	return middleware.RealIP(next)
}

func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return middleware.Timeout(timeout)
}
