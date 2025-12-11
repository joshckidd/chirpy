package main

import (
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("GET /api/healthz", readinessEndpoint)
	server := http.Server{
		Handler: serveMux,
		Addr:    ":8080",
	}

	var apiCfg apiConfig
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir("/home/josh/Documents/repos/github.com/joshckidd/chirpy")))))
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.returnMetrics)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)
	serveMux.HandleFunc("POST /api/validate_chirp", validateChirp)
	server.ListenAndServe()
}
