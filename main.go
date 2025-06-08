package main

import (
	"findservers/api"
	"findservers/cache"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cache.InitRedis()
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)
	http.HandleFunc("/api/servers", api.GetServers)
	http.HandleFunc("/api/refresh-servers", api.RefreshServers)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
