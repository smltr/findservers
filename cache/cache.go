package cache

import (
	"context"
	"encoding/json"
	"findservers/models"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	rdb           *redis.Client
	ctx           = context.Background()
	serversKey    = "servers:list"
	cacheDuration = 15 * time.Minute
)

func InitRedis() error {
	redisURL := os.Getenv("REDIS_URL")

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	rdb = redis.NewClient(opt)

	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Connected to Redis successfully")
	return nil
}

func RefreshServerCache() {
	servers, err := fetchServers()
	if err != nil {
		log.Printf("Failed to fetch servers: %v", err)
		return
	}

	jsonData, err := json.Marshal(servers)
	if err != nil {
		log.Printf("Failed to marshal servers: %v", err)
		return
	}

	err = rdb.Set(ctx, serversKey, jsonData, cacheDuration).Err()
	if err != nil {
		log.Printf("Failed to cache servers in Redis: %v", err)
		return
	}

	log.Printf("Successfully cached %d servers in Redis (expires in %v)", len(servers), cacheDuration)
}

func GetServersFromCache() []models.Server {
	jsonData, err := rdb.Get(ctx, serversKey).Result()
	if err != nil {
		if err == redis.Nil {
			log.Println("No servers in cache, fetching fresh data...")
			RefreshServerCache()

			jsonData, err = rdb.Get(ctx, serversKey).Result()
			if err != nil {
				log.Printf("Failed to get servers from cache after refresh: %v", err)
				return []models.Server{}
			}
		} else {
			log.Printf("Redis error: %v", err)
			return []models.Server{}
		}
	}

	var servers []models.Server
	err = json.Unmarshal([]byte(jsonData), &servers)
	if err != nil {
		log.Printf("Failed to unmarshal servers from cache: %v", err)
		return []models.Server{}
	}

	return servers
}

func GetCacheInfo() (bool, time.Duration, int, error) {
	ttl, err := rdb.TTL(ctx, serversKey).Result()
	if err != nil {
		return false, 0, 0, err
	}

	exists := ttl > 0
	var count int

	if exists {
		servers := GetServersFromCache()
		count = len(servers)
	}

	return exists, ttl, count, nil
}

func CloseRedis() error {
	if rdb != nil {
		return rdb.Close()
	}
	return nil
}

func fetchServers() ([]models.Server, error) {
	apiKey := os.Getenv("STEAM_API_KEY")

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

	var rawServers []models.ServerRaw
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
					Servers []models.ServerRaw `json:"servers"`
				} `json:"response"`
			}

			if err := json.Unmarshal(rawBody, &result); err != nil {
				log.Printf("Region %s, Attempt %d: JSON parse error: %v", region, i+1, err)
				continue
			}

			rawServers = append(rawServers, result.Response.Servers...)

			if len(result.Response.Servers) > 0 {
				break
			}

		}
	}

	servers := FilterAndCleanServers(rawServers)
	log.Printf("Raw servers found across all regions: %d", len(rawServers))
	log.Printf("Total servers found across all regions after filtering: %d", len(servers))
	return servers, nil

}

func FilterAndCleanServers(rawServers []models.ServerRaw) []models.Server {
	foundIPs := make(map[string]bool)
	servers := make([]models.Server, 0)

	for _, server := range rawServers {
		// Skip Valve official servers
		switch {
		case strings.HasPrefix(server.Name, "Valve Counter-Strike"):
			continue
		case server.MaxPlayers > 64: // based on experience, anything with 255 max players is spam
			continue
		case strings.Contains(server.GameType, "stalnoy"):
			continue
		case foundIPs[server.Addr]:
			continue
		case server.Map == "":
			continue
		}

		foundIPs[server.Addr] = true
		servers = append(servers, models.CleanServer(server))
	}

	return servers
}
