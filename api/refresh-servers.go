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

func RefreshServers(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("STEAM_API_KEY")
	// get blob url from environment variable
	blobURL := os.Getenv("BLOB_URL")
	if apiKey == "" || blobURL == "" {
		http.Error(w, "STEAM_API_KEY environment variable not set", http.StatusInternalServerError)
		return
	}

	servers, err := fetchServers(apiKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// marshal servers to json
	jsonBytes, err := json.Marshal(servers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonReader := bytes.NewReader(jsonBytes)

	// save servers to vercel blob storage directly via HTTP
	blobURL += "/servers.json"
	apiToken := os.Getenv("BLOB_READ_WRITE_TOKEN")

	if apiToken == "" {
		http.Error(w, "BLOB_READ_WRITE_TOKEN environment variable not set", http.StatusInternalServerError)
		return
	}

	// Create a new request
	req, err := http.NewRequest("PUT", blobURL, jsonReader)
	if err != nil {
		fmt.Println("Error creating request:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add headers
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+apiToken)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error uploading to blob storage:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check if response is successful
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("Failed to upload blob: status code %d, response: %s", resp.StatusCode, string(respBody))
		fmt.Println(errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	// print success message
	fmt.Println("Servers refreshed successfully")
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
		for i := 0; i < maxRetries; i++ {
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

			// Filter and process servers
			for _, server := range result.Response.Servers {
				// Skip Valve official servers
				if strings.HasPrefix(server.Name, "Valve Counter-Strike") {
					continue
				}

				allServers = append(allServers, server)
			}

			// If we got servers, break the retry loop for this region
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
