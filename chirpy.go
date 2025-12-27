package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joshckidd/chirpy/internal/auth"
	"github.com/joshckidd/chirpy/internal/database"
)

type returnUserRow struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

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
	type userParam struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type returnUserRow struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}

	decoder := json.NewDecoder(r.Body)
	inParams := userParam{}

	err := decoder.Decode(&inParams)
	if err != nil {
		respondWithError(w, 500, "Invalid request")
		return
	}

	hashedPassword, err := auth.HashPassword(inParams.Password)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	params := database.CreateUserParams{
		Email:          inParams.Email,
		HashedPassword: hashedPassword,
	}

	user, err := cfg.db.CreateUser(r.Context(), params)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 201, returnUserRow{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed.Bool,
	})
}

func (cfg *apiConfig) postChirp(w http.ResponseWriter, r *http.Request) {
	type chirpParam struct {
		Body string `json:"body"`
	}

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	id, err := auth.ValidateJWT(tokenString, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := chirpParam{}
	err = decoder.Decode(&params)
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
		UserID: id,
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), createParams)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 201, chirp)
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	var chirps []database.Chirp

	s := r.URL.Query().Get("sort")
	a := r.URL.Query().Get("author_id")
	if a != "" {
		id, err := uuid.Parse(a)
		if err != nil {
			respondWithError(w, 500, err.Error())
			return
		}
		chirps, err = cfg.db.GetChirpsForUser(r.Context(), id)
		if err != nil {
			respondWithError(w, 500, err.Error())
			return
		}

	} else {
		var err error
		chirps, err = cfg.db.GetAllChirps(r.Context())
		if err != nil {
			respondWithError(w, 500, err.Error())
			return
		}
	}

	if s == "desc" {
		sort.Slice(chirps, func(i, j int) bool { return chirps[i].CreatedAt.After(chirps[j].CreatedAt) })
	}

	respondWithJSON(w, 200, chirps)
}

func (cfg *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	chirp, err := cfg.db.GetChirp(r.Context(), id)
	if err != nil {
		respondWithError(w, 404, "Chirp not found")
		return
	}

	respondWithJSON(w, 200, chirp)
}

func (cfg *apiConfig) userLogin(w http.ResponseWriter, r *http.Request) {
	type loginParam struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type returnUserRow struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
	}

	decoder := json.NewDecoder(r.Body)
	params := loginParam{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Invalid request")
		return
	}

	user, err := cfg.db.GetUserWithEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	tok, err := auth.MakeJWT(user.ID, cfg.tokenSecret, time.Hour)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	val, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if val == true {
		randTok, _ := auth.MakeRefreshToken()
		rt, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
			UserID: user.ID,
			Token:  randTok,
		})
		if err != nil {
			respondWithError(w, 500, err.Error())
			return
		}
		userResp := returnUserRow{
			ID:           user.ID,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
			Email:        user.Email,
			Token:        tok,
			RefreshToken: rt.Token,
			IsChirpyRed:  user.IsChirpyRed.Bool,
		}
		respondWithJSON(w, 200, userResp)
		return
	}
	respondWithError(w, 401, "Incorrect email or password")
}

func (cfg *apiConfig) refreshJWT(w http.ResponseWriter, r *http.Request) {
	type returnUserRow struct {
		Token string `json:"token"`
	}

	tok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	userID, err := cfg.db.GetUserFromRefreshToken(r.Context(), tok)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	accessToken, err := auth.MakeJWT(userID, cfg.tokenSecret, time.Hour)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 200, returnUserRow{Token: accessToken})
}

func (cfg *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), tok)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 204, nil)
}

func (cfg *apiConfig) putUser(w http.ResponseWriter, r *http.Request) {
	type userParam struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	id, err := auth.ValidateJWT(tokenString, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	decoder := json.NewDecoder(r.Body)
	inParams := userParam{}

	err = decoder.Decode(&inParams)
	if err != nil {
		respondWithError(w, 500, "Invalid request")
		return
	}

	hashedPassword, err := auth.HashPassword(inParams.Password)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	params := database.UpdateUserParams{
		Email:          inParams.Email,
		HashedPassword: hashedPassword,
		ID:             id,
	}

	user, err := cfg.db.UpdateUser(r.Context(), params)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	respondWithJSON(w, 200, user)
}

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	id, err := auth.ValidateJWT(tokenString, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	c, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}
	if c.UserID != id {
		respondWithError(w, 403, "Unauthorized user")
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}

	respondWithJSON(w, 204, nil)
}

func (cfg *apiConfig) userRed(w http.ResponseWriter, r *http.Request) {
	type dataParams struct {
		UserID uuid.UUID `json:"user_id"`
	}

	type redParams struct {
		Event string     `json:"event"`
		Data  dataParams `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != cfg.polkaKey {
		respondWithError(w, 401, "Invalid API key")
		return
	}

	decoder := json.NewDecoder(r.Body)
	inParams := redParams{}

	err = decoder.Decode(&inParams)
	if err != nil {
		respondWithError(w, 500, "Invalid request")
		return
	}

	if inParams.Event != "user.upgraded" {
		respondWithJSON(w, 204, nil)
		return
	}

	err = cfg.db.UpdateUserRed(r.Context(), inParams.Data.UserID)
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}

	respondWithJSON(w, 204, nil)
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
