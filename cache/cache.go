package cache

import (
	"encoding/json"
	"findservers/models"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

var (
	serverCache = []models.Server{}
	lock        = sync.RWMutex{}
)

func RefreshServerCache() {
	servers, err := fetchServers()
	if err != nil {
		log.Printf("Failed to fetch servers: %v", err)
		return
	}

	lock.Lock()
	defer lock.Unlock()

	serverCache = servers
}

func GetServersFromCache() []models.Server {
	lock.RLock()
	defer lock.RUnlock()

	return serverCache
}

func fetchServers() ([]models.Server, error) {
	apiKey := os.Getenv("STEAM_API_KEY")
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
					Servers []models.ServerRaw `json:"servers"`
				} `json:"response"`
			}

			if err := json.Unmarshal(rawBody, &result); err != nil {
				log.Printf("Region %s, Attempt %d: JSON parse error: %v", region, i+1, err)
				continue
			}

			// some basic preliminary filtering
			for _, server := range result.Response.Servers {
				// Skip Valve official servers
				if strings.HasPrefix(server.Name, "Valve Counter-Strike") {
					continue
				}

				allServers = append(allServers, models.CleanServer(server))
			}

			if len(result.Response.Servers) > 0 {
				break
			}

		}
	}

	log.Printf("Total servers found across all regions: %d", len(allServers))

	if len(allServers) > 10 {
		return allServers, nil
	}

	return nil, fmt.Errorf("failed to fetch sufficient servers across all regions")
}
