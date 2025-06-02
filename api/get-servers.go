package api

import (
	"encoding/json"
	"net/http"

	"findservers/cache"
)

func GetServers(w http.ResponseWriter, r *http.Request) {
	servers := cache.GetServersFromCache()
	jsonData, err := json.Marshal(servers)
	if err != nil {
		http.Error(w, "Failed to marshal servers to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
