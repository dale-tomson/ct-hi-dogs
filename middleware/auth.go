package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

func APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedKey := os.Getenv("API_KEY")
		if expectedKey == "" {
			slog.Error("API_KEY environment variable is not configured")
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "server misconfiguration: API_KEY not set",
			})
			return
		}

		provided := r.Header.Get("X-API-Key")
		if provided == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "missing X-API-Key header",
			})
			return
		}
		if provided != expectedKey {
			slog.Warn("invalid API key", "remote_addr", r.RemoteAddr)
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid API key",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
