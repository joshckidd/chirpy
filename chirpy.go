package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joshckidd/chirpy/internal/database"
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

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	if cfg.environment != "dev" {
		w.WriteHeader(403)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
	cfg.fileserverHits.Store(0)
	cfg.db.ResetUsers(r.Context())
}

func (cfg *apiConfig) postUser(w http.ResponseWriter, r *http.Request) {
	type emailParam struct {
		Email string `json:"email"`
	}

	type returnUser struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := emailParam{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Invalid request")
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), params.Email)

	resp := returnUser{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, 201, resp)
}

func (cfg *apiConfig) postChirp(w http.ResponseWriter, r *http.Request) {
	type chirpParam struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	type returnChirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := chirpParam{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Invalid request")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	createParams := database.CreateChirpParams{
		Body:   cleanChirp(params.Body),
		UserID: params.UserID,
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), createParams)

	resp := returnChirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, 201, resp)
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
