package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func GetServers(w http.ResponseWriter, r *http.Request) {
	println("getting servers...")

	blobURL := os.Getenv("BLOB_URL") + "/servers.txt"

	req, err := http.NewRequest("GET", blobURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	apiKey := os.Getenv("BLOB_READ_WRITE_TOKEN")
	if apiKey != "" {
		req.Header.Add("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("Failed to fetch blob: status code %d", resp.StatusCode)
		fmt.Println(errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	htmlContent, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	w.WriteHeader(http.StatusOK)
	w.Write(htmlContent)
}
