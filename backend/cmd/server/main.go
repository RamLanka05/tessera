package main

import (
	"fmt"
	"net/http"
	
)

func main() {
	fmt.Println("Starting Tessera High-Performance Engine...")

	mux := http.NewServeMux()

	http.HandleFunc("/api/v1/config", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "healthy"}`))
	})

	handler := authMiddleware(mux)

	http.ListenAndServe(":8080", handler)
}