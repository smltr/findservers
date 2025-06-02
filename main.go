package main

import (
	"findservers/api"
	"findservers/cache"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cache.RefreshServerCache()
	go func() {
		for {
			time.Sleep(time.Minute)
			cache.RefreshServerCache()
		}
	}()
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)
	http.HandleFunc("/api/servers", api.GetServers)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
