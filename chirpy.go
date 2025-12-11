package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func readinessEndpoint(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) returnMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
	cfg.fileserverHits.Store(0)
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}

	type returnError struct {
		Error string `json:"error"`
	}

	type returnValid struct {
		Valid bool `json:"valid"`
	}

	respGenError := returnError{
		Error: "Something went wrong",
	}

	dat, err := json.Marshal(respGenError)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	decoder := json.NewDecoder(r.Body)
	c := chirp{}
	err = decoder.Decode(&c)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write(dat)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(c.Body) > 140 {
		respError := returnError{
			Error: "Chirp is too long",
		}

		long, err := json.Marshal(respError)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			w.Write(dat)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(long)
		return
	} else {
		respError := returnValid{
			Valid: true,
		}

		val, err := json.Marshal(respError)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			w.Write(dat)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(val)
		return
	}
}
