package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"findservers/models"

	"github.com/rpdg/vercel_blob"
)

func RefreshServers(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("STEAM_API_KEY")
	blobURL := os.Getenv("BLOB_URL")
	if apiKey == "" || blobURL == "" {
		http.Error(w, "STEAM_API_KEY environment variable not set", http.StatusInternalServerError)
		return
	}

	servers, err := fetchServers(apiKey)
	var filteredServers []models.Server
	for _, server := range servers {
		if server.Region == 0 || server.Region == 1 {
			filteredServers = append(filteredServers, server)
		}
	}
	servers = filteredServers
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	text := generateServerHTML(servers)

	blobURL += "/servers.txt"

	println(blobURL)

	apiToken := os.Getenv("BLOB_READ_WRITE_TOKEN")

	if apiToken == "" {
		http.Error(w, "BLOB_READ_WRITE_TOKEN environment variable not set", http.StatusInternalServerError)
		return
	}

	client := vercel_blob.NewVercelBlobClient()
	result, err := client.Put(
		"servers.txt",
		strings.NewReader(text),
		vercel_blob.PutCommandOptions{
			AddRandomSuffix: false,
			ContentType:     "text/plain",
		},
	)

	if err != nil {
		fmt.Println("Error uploading to blob storage:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Servers refreshed successfully, %d total, uploaded to: %s\n",
		len(servers),
		result.URL,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"url":    result.URL,
		"count":  strconv.Itoa(len(servers)),
	})
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

func generateServerHTML(servers []models.Server) string {
	var html strings.Builder

	html.WriteString(fmt.Sprintf(`<span id="server-count" hx-swap-oob="true">%d</span>`, len(servers)))

	addrsAlreadyAdded := make(map[string]bool)

	for i, server := range servers {
		if addrsAlreadyAdded[server.Addr] {
			continue
		}
		addrsAlreadyAdded[server.Addr] = true
		secure := ""
		if server.Secure {
			secure = "●"
		}
		bots := ""
		if server.Bots > 0 {
			bots = fmt.Sprintf("%d", server.Bots)
		}
		html.WriteString(fmt.Sprintf(`
            <div
				id="%d"
                class="server-row"
                :class="{ 'selected': selectedServer === '%s' }"
                @click="selectedServer = '%s'"
                @dblclick="window.location.href = 'steam://connect/%s'">
                <div class="cell-pw"></div>
                <div class="cell-vac">%v</div>
                <div class="cell-region">%s</div>
                <div class="cell-name">
                    <span class="server-name">%s</span>
                    <span class="server-ip">%s</span>
                </div>
                <div class="cell-bot">%s</div>
                <div class="cell-players">%s</div>
                <div class="cell-map">%s</div>
                <div class="cell-tags">%s</div>
            </div>
            `,
			i,
			server.Addr,
			server.Addr,
			server.Addr,
			secure,
			regionCodeToString(server.Region),
			template.HTMLEscapeString(server.Name),
			server.Addr,
			bots,
			formatPlayers(server.Players, server.MaxPlayers),
			template.HTMLEscapeString(server.Map),
			template.HTMLEscapeString(server.GameType),
		))
	}
	return html.String()
}

func regionCodeToString(code int) string {
	regions := map[int]string{
		0: "US", 1: "US", 2: "SA", 3: "EU",
		4: "AS", 5: "AU", 6: "ME", 7: "AF",
		255: "WD",
	}
	if reg, ok := regions[code]; ok {
		return reg
	}
	return "??"
}

func formatPlayers(players, maxPlayers int) string {
	if players == 0 {
		return fmt.Sprintf("<span class='zero-players'>%d</span><span class='max-players'>/%d</span>",
			players, maxPlayers)
	}
	return fmt.Sprintf("%d<span class='max-players'>/%d</span>", players, maxPlayers)
}
