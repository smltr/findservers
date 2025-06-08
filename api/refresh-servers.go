package api

import (
	"net/http"
	"os"
	"strings"

	"findservers/cache"
)

func RefreshServers(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	expectedToken := os.Getenv("CRON_API_TOKEN")
	if token != expectedToken {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	err := cache.RefreshServerCache()
	if err != nil {
		http.Error(w, "Failed to refresh server cache", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte{})
}
