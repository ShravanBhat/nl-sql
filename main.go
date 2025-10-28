package main

import (
	"log"
	"net/http"
	"nlsql/api"

	"github.com/joho/godotenv"
)

func main() {
	loadAIKey()
	initHTTP()
}

func loadAIKey() {
	// Load .env file at startup
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Could not load .env file. Proceeding with environment variables.")
	}

}

func initHTTP() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	// API endpoints
	http.HandleFunc("/api/connect", api.HandleConnect)
	http.HandleFunc("/api/generate-query", api.HandleGenerateQuery)
	http.HandleFunc("/api/execute-query", api.HandleExecuteQuery)

	log.Println("Starting Go SQL Agent server")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
