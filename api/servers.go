package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"findservers/models"

	"github.com/redis/go-redis/v9"
)

type CachedData struct {
	Servers   []models.Server `json:"servers"`
	Timestamp time.Time       `json:"timestamp"`
}

func GetServers(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("STEAM_API_KEY")
	if apiKey == "" {
		http.Error(w, "STEAM_API_KEY environment variable not set", http.StatusInternalServerError)
		return
	}

	// TODO make this work, currently just fetches new servers each time
	// Try to get cached data first
	cachedData, err := getCachedServers()
	if err == nil && time.Since(cachedData.Timestamp) < 5*time.Minute {
		log.Printf("Using cached data from %v", cachedData.Timestamp)
		jsonBytes, err := json.Marshal(cachedData.Servers)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
		return
	}

	// Cache is stale or doesn't exist, fetch new data
	log.Printf("Cache miss or stale, fetching new server data")
	servers, err := fetchServers(apiKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// encode servers to json and write to request
	jsonBytes, err := json.Marshal(servers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)

	// Save to cache asynchronously
	go func() {
		if err := saveCachedServers(servers); err != nil {
			log.Printf("Failed to save cache: %v", err)
		}
	}()
}

func fetchServers(apiKey string) ([]models.Server, error) {
	var allServers []models.Server
	maxRetries := 3
	// Try different regions to split up the results and avoid the 10k limit
	regions := []string{
		"\\region\\0",   // US East
		"\\region\\1",   // US West
		"\\region\\2",   // South America
		"\\region\\3",   // Europe
		"\\region\\4",   // Asia
		"\\region\\5",   // Australia
		"\\region\\6",   // Middle East
		"\\region\\7",   // Africa
		"\\region\\255", // Worldwide
	}

	for _, region := range regions {
		for i := range maxRetries {
			// Basic filter for dedicated servers and specific region
			url := fmt.Sprintf("https://api.steampowered.com/IGameServersService/GetServerList/v1/?key=%s&filter=\\appid\\730\\dedicated\\1%s&limit=10000", apiKey, region)

			resp, err := http.Get(url)
			if err != nil {
				log.Printf("Region %s, Attempt %d: HTTP error: %v", region, i+1, err)
				continue
			}

			log.Printf("Region %s, Attempt %d: Status Code: %d", region, i+1, resp.StatusCode)

			rawBody, err := io.ReadAll(resp.Body)
			resp.Body.Close()

			if err != nil {
				log.Printf("Region %s, Attempt %d: Body read error: %v", region, i+1, err)
				continue
			}

			var result struct {
				Response struct {
					Servers []models.Server `json:"servers"`
				} `json:"response"`
			}

			if err := json.Unmarshal(rawBody, &result); err != nil {
				log.Printf("Region %s, Attempt %d: JSON parse error: %v", region, i+1, err)
				continue
			}

			for _, server := range result.Response.Servers {
				// Skip Valve official servers
				if strings.HasPrefix(server.Name, "Valve Counter-Strike") {
					continue
				}

				allServers = append(allServers, server)
			}

			if len(result.Response.Servers) > 0 {
				break
			}

			time.Sleep(time.Second * 2)
		}
	}

	log.Printf("Total servers found across all regions: %d", len(allServers))

	if len(allServers) > 10 {
		return allServers, nil
	}

	return nil, fmt.Errorf("failed to fetch sufficient servers across all regions")
}

func getCachedServers() (*CachedData, error) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL environment variable not set")
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %v", err)
	}

	rdb := redis.NewClient(opt)
	defer rdb.Close()

	ctx := context.Background()
	data, err := rdb.Get(ctx, "servers_cache").Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("cache not found")
	} else if err != nil {
		return nil, fmt.Errorf("redis error: %v", err)
	}

	var cachedData CachedData
	if err := json.Unmarshal([]byte(data), &cachedData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached data: %v", err)
	}

	return &cachedData, nil
}

func saveCachedServers(servers []models.Server) error {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return fmt.Errorf("REDIS_URL environment variable not set")
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("failed to parse Redis URL: %v", err)
	}

	rdb := redis.NewClient(opt)
	defer rdb.Close()

	cachedData := CachedData{
		Servers:   servers,
		Timestamp: time.Now(),
	}

	jsonData, err := json.Marshal(cachedData)
	if err != nil {
		return fmt.Errorf("failed to marshal cached data: %v", err)
	}

	ctx := context.Background()
	err = rdb.Set(ctx, "servers_cache", jsonData, 10*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("failed to save to Redis: %v", err)
	}

	log.Printf("Successfully cached server data in Redis")
	return nil
}
