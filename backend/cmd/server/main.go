package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"tessera/backend/internal/auth"
	"tessera/backend/internal/storage"
)

type AuthTokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type AuthTokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
}

type ErrorResponse struct {
	Message string `json:"error"`
}

type CreateConfigRequest struct {
	ConfigKey   string      `json:"config_key"`
	ConfigValue interface{} `json:"config_val"`
	Author      string      `json:"author"`
	Message     string      `json:"message"`
}

type CreateConfigResponse struct {
	VersionID int `json:"version_id"`
}

type GetConfigResponse struct {
	ConfigValue interface{} `json:"config_val"`
}

type GetVersionsResponse struct {
	Versions []storage.Version `json:"versions"`
}

type RollbackResponse struct {
	VersionID int `json:"version_id"`
}

var db *sql.DB

func authTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	fmt.Printf("authTokenHandler raw body: %s\n", string(body))
	r.Body = io.NopCloser(bytes.NewReader(body))

	var req AuthTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("authTokenHandler decode error: %v\n", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	valid, err := storage.ValidateClient(db, req.ClientID, req.ClientSecret)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal server error: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Printf("authTokenHandler validation result: client_id=%s valid=%t\n", req.ClientID, valid)

	if !valid {
		http.Error(w, "Invalid client credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateJWT(req.ClientID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate token: %v", err), http.StatusInternalServerError)
		return
	}

	response := AuthTokenResponse{
		Token:     token,
		ExpiresIn: 15 * 60, // 15 minutes = 900 seconds
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleCreateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.ConfigKey == "" || req.ConfigValue == nil || req.Author == "" || req.Message == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	currVID, err := storage.CreateConfig(db, req.ConfigKey, req.ConfigValue, req.Author, req.Message)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create config: %v", err), http.StatusInternalServerError)
		return
	}

	response := CreateConfigResponse{
		VersionID: currVID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	configVal, err := storage.GetConfig(db, name)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get config: %v", err), http.StatusNotFound)
		return
	}

	// Unmarshal the []byte into an interface{}
	var val interface{}
	if err := json.Unmarshal(configVal, &val); err != nil {
		http.Error(w, fmt.Sprintf("Failed to unmarshal config value: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetConfigResponse{ConfigValue: val})
}

func handleListVersions(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	allVers, err := storage.ListVersions(db, name)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list versions: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"versions": allVers})
}

func handleRollbackConfig(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	vidStr := r.PathValue("vid")
	
	var vid int
	_, err := fmt.Sscanf(vidStr, "%d", &vid)
	if err != nil {
		http.Error(w, "Invalid version ID", http.StatusBadRequest)
		return
	}
	
	newVID, err := storage.RollbackConfig(db, name, vid)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to rollback: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"version_id": newVID})
}

func main() {

	fmt.Println("Starting Tessera High-Performance Engine...")

	db = storage.InitDB()
	defer db.Close()

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		fmt.Println("Migrations completed. Exiting.")
		return
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/auth/token", authTokenHandler)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message": "Welcome to Tessera!"}`))
	})

	mux.Handle("/api/v1/config", auth.AuthMiddleware(http.HandlerFunc(handleCreateConfig)))

	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
