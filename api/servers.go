package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"findservers/models"
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
	// go func() {
	if err := saveCachedServers(servers); err != nil {
		log.Printf("Failed to save cache: %v", err)
	}
	// }()
}

func fetchServers(apiKey string) ([]models.Server, error) {
	var allServers []models.Server
	maxRetries := 3
	// Try different regions to split up the results and avoid the 10k limit
	//
	regions := []string{
		"\\region\\0", // US East
		"\\region\\1", // US West
		"\\region\\2", // South America
		"\\region\\3", // Europe
		// "\\region\\4",   // Asia
		// "\\region\\5",   // Australia
		// "\\region\\6",   // Middle East
		// "\\region\\7",   // Africa
		// "\\region\\255", // Worldwide
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
	blobURL := os.Getenv("BLOB_URL")
	blobToken := os.Getenv("BLOB_READ_WRITE_TOKEN")
	if blobURL == "" || blobToken == "" {
		return nil, fmt.Errorf("BLOB_URL or BLOB_READ_WRITE_TOKEN not set")
	}

	// Try to fetch servers.json from blob storage
	req, err := http.NewRequest("GET", blobURL+"/servers.json", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+blobToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("blob not found or error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cachedData CachedData
	if err := json.Unmarshal(body, &cachedData); err != nil {
		return nil, err
	}

	return &cachedData, nil
}

func saveCachedServers(servers []models.Server) error {
	blobToken := os.Getenv("BLOB_READ_WRITE_TOKEN")
	if blobToken == "" {
		return fmt.Errorf("BLOB_READ_WRITE_TOKEN not set")
	}

	cachedData := CachedData{
		Servers:   servers,
		Timestamp: time.Now(),
	}

	jsonData, err := json.Marshal(cachedData)
	if err != nil {
		return err
	}

	// PUT to blob storage using proper API endpoint
	req, err := http.NewRequest("PUT", "https://blob.vercel-storage.com/servers.json", bytes.NewReader(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+blobToken)
	req.Header.Set("x-api-version", "9")
	req.Header.Set("X-Add-Random-Suffix", "0") // Don't add random suffix, we want consistent filename
	req.Header.Set("X-Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to save to blob: %d - %s", resp.StatusCode, string(body))
	}

	log.Printf("Successfully cached server data")
	return nil
}
