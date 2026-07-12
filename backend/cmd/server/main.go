package main

import (
	"fmt"
	"net/http"
 	"tessera/backend/internal/auth"
	"tessera/backend/internal/storage"

	
)

func main() {

	fmt.Println("Starting Tessera High-Performance Engine...")

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message": "Welcome to Tessera!"}`))
	})


	mux.Handle("/api/v1/config", auth.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "healthy"}`))
	})))


	storage.InitDB()
	
	http.ListenAndServe(":8080", mux)
}