package main

import (
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// Load environment variables
	envPath := filepath.Join("../../", ".env")
	err := godotenv.Load(envPath)
	if err != nil {
		log.Warn().Msg("No .env file found, using system environment variables")
	}
	// Configure Zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Read port from environment
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}


	// Initialize Router
	router := mux.NewRouter()

	// Define Routes
	router.HandleFunc("/", HomeHandler).Methods("GET")

	// Start Server
	log.Info().Msgf("Server running on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}

// HomeHandler serves the root endpoint
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Home endpoint hit")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Facebook Creatives Platform is running"))
}
