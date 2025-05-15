package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	// Parse optional file flag
	filePath := flag.String("file", "", "Path to JSON file to send in POST request")
	flag.Parse()

	var body []byte
	var err error

	// Read file content if file path is provided
	if *filePath != "" {
		body, err = os.ReadFile(*filePath)
		if err != nil {
			log.Fatalf("Error reading file: %v", err)
		}
	} else {
		body = []byte(`{}`)
	}

	// Set up request
	req, err := http.NewRequest("POST", "https://localhost:8085/kamar-refresh", bytes.NewReader(body))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+basicAuth("admin", "asdfasdf"))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Print response
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %s\nResponse: %s\n", resp.Status, string(respBody))
}

// basicAuth generates the Basic Auth header value
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
