package rest

import (
	"crypto/subtle"
	"net/http"
)

func authMiddleware(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("X-Engine-Token")
		if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			writeEngineError(w, http.StatusUnauthorized, "unauthorized", "missing or invalid token", false)
			return
		}
		next.ServeHTTP(w, r)
	})
}
