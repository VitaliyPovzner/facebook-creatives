package main

import (
	"facebook-creatives/internal/service"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)



func main() {
	initEnv()
	initLogger()

	port := getEnv("APP_PORT", "8080")
	accessToken := os.Getenv("FB_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal().Msg("FB_ACCESS_TOKEN is not set")
	}

	facebookService := service.NewFacebookService(accessToken)
	// ctx:=context.TODO()
	go facebookService.FetchCreativeDataPipeline()

	router := initRouter()
	log.Info().Msgf("Server running on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}

func initEnv() {
	envPath := filepath.Join("../../", ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Warn().Msg("No .env file found, using system environment variables")
	}
}

func initLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func initRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", HomeHandler).Methods("GET")
	return router
}


func HomeHandler(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Home endpoint hit")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Facebook Creatives Platform is running"))
}
