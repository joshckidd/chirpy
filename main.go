package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	"github.com/joshckidd/chirpy/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	environment    string
	tokenSecret    string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("Database error.")
		os.Exit(1)
	}

	dbQueries := database.New(db)

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("GET /api/healthz", readinessEndpoint)
	server := http.Server{
		Handler: serveMux,
		Addr:    ":8080",
	}

	var apiCfg apiConfig
	apiCfg.db = dbQueries
	apiCfg.environment = os.Getenv("PLATFORM")
	apiCfg.tokenSecret = os.Getenv("SECRET")
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir("/home/josh/Documents/repos/github.com/joshckidd/chirpy")))))
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.returnMetrics)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)
	serveMux.HandleFunc("POST /api/users", apiCfg.postUser)
	serveMux.HandleFunc("POST /api/chirps", apiCfg.postChirp)
	serveMux.HandleFunc("GET /api/chirps", apiCfg.getChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirp)
	serveMux.HandleFunc("POST /api/login", apiCfg.userLogin)
	serveMux.HandleFunc("POST /api/refresh", apiCfg.refreshJWT)
	serveMux.HandleFunc("POST /api/revoke", apiCfg.revokeToken)
	server.ListenAndServe()
}
