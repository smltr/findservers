package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"findservers/models"
)

func GetServers(w http.ResponseWriter, r *http.Request) {
	println("getting servers...")

	// Direct URL to the blob
	blobURL := os.Getenv("BLOB_URL") + "/servers.json"

	// Create a new request
	req, err := http.NewRequest("GET", blobURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add auth token if needed
	apiKey := os.Getenv("BLOB_READ_WRITE_TOKEN")
	if apiKey != "" {
		req.Header.Add("Authorization", "Bearer "+apiKey)
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check if response is successful
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("Failed to fetch blob: status code %d", resp.StatusCode)
		fmt.Println(errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Read the response body
	jsonBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	println("unmarshalling...")
	// Create blob into io reader
	reader := bytes.NewReader(jsonBytes)

	// try to unmarshal to json
	var servers []models.Server
	if err := json.NewDecoder(reader).Decode(&servers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// write response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servers)
}
