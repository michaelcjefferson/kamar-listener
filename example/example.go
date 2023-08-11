package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type UploadData struct {
	Sync bool `json:"sync"`
}

func checkBasicAuth(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}

	authParts := strings.Split(authHeader, " ")
	if len(authParts) != 2 || authParts[0] != "Basic" {
		return false
	}

	decodedAuth, err := base64.StdEncoding.DecodeString(authParts[1])
	if err != nil {
		return false
	}

	authCredentials := strings.Split(string(decodedAuth), ":")
	if len(authCredentials) != 2 || authCredentials[0] != "David" || authCredentials[1] != "pa55word" {
		return false
	}

	return true
}

func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check HTTP Basic Authorization
	if !checkBasicAuth(r) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Authorization Required"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the multipart form data
	err := r.ParseMultipartForm(10 << 20) // 10 MB max file size
	if err != nil {
		http.Error(w, "Unable to parse form data", http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, _, err := r.FormFile("file") // Assuming the field name in the form is "file"
	if err != nil {
		http.Error(w, "Unable to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read the uploaded file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file data", http.StatusInternalServerError)
		return
	}

	// Unmarshal JSON data to check for "sync" field
	var uploadData UploadData
	err = json.Unmarshal(fileData, &uploadData)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Check if "sync" field is present and true
	if !uploadData.Sync {
		http.Error(w, `"sync" field is missing or false`, http.StatusBadRequest)
		return
	}

	// Create a new JSON file on the server
	outputFile, err := os.Create("uploaded.json")
	if err != nil {
		http.Error(w, "Unable to create file on server", http.StatusInternalServerError)
		return
	}
	defer outputFile.Close()

	// Write the uploaded file data to the server's JSON file
	_, err = outputFile.Write(fileData)
	if err != nil {
		http.Error(w, "Error writing file data", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File uploaded successfully!")
}

func main() {
	http.HandleFunc("/upload", handleFileUpload)
	port := 8080
	fmt.Printf("Server started on port %d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}
