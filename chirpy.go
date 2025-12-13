package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

	type returnValid struct {
		Valid       bool   `json:"valid"`
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	c := chirp{}
	err := decoder.Decode(&c)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(c.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	} else {
		resp := returnValid{
			Valid:       true,
			CleanedBody: cleanChirp(c.Body),
		}

		respondWithJSON(w, 200, resp)

		return
	}
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type returnError struct {
		Error string `json:"error"`
	}

	respError := returnError{
		Error: msg,
	}

	dat, err := json.Marshal(respError)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	val, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(val)
}

func cleanChirp(c string) string {
	words := strings.Split(c, " ")

	for i := range words {
		if strings.ToLower(words[i]) == "kerfuffle" ||
			strings.ToLower(words[i]) == "sharbert" ||
			strings.ToLower(words[i]) == "fornax" {
			words[i] = "****"
		}
	}

	res := strings.Join(words, " ")

	return res
}
