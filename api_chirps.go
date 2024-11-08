package main

import (
	"encoding/json"
	"net/http"
	"time"

	"strings"

	"github.com/KidMuon/chirpy/internal/auth"
	"github.com/KidMuon/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handleGetAllChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	chirps := []Chirp{}
	for _, dbChrip := range dbChirps {
		chirps = append(chirps, dbChirpToChirp(dbChrip))
	}

	respondWithJSON(w, 200, chirps)
}

func (cfg *apiConfig) handleGetChripByID(w http.ResponseWriter, r *http.Request) {
	chirpUUID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	dbChirp, err := cfg.db.GetChirpByID(r.Context(), chirpUUID)
	if err != nil {
		respondWithError(w, 404, "Not Found")
		return
	}
	respondWithJSON(w, 200, dbChirpToChirp(dbChirp))

}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type requestChirp struct {
		Body    string    `json:"body"`
		User_ID uuid.UUID `json:"user_id"`
	}

	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	var reqChirp requestChirp
	err := decoder.Decode(&reqChirp)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	var chirpCharacterCount int
	for range reqChirp.Body {
		chirpCharacterCount++
	}

	if chirpCharacterCount > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	token_ID, err := auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	reqChirp.User_ID = token_ID

	chirpToCreate := database.CreateChirpParams{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      getCleanedChirpBody(reqChirp.Body),
		UserID:    reqChirp.User_ID,
	}
	dbChirp, err := cfg.db.CreateChirp(r.Context(), chirpToCreate)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 201, dbChirpToChirp(dbChirp))
}

type Chirp struct {
	ID         uuid.UUID `json:"id"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
	Body       string    `json:"body"`
	User_ID    uuid.UUID `json:"user_id"`
}

func dbChirpToChirp(dbChirp database.Chirp) Chirp {
	return Chirp{
		ID:         dbChirp.ID,
		Created_at: dbChirp.CreatedAt,
		Updated_at: dbChirp.UpdatedAt,
		Body:       dbChirp.Body,
		User_ID:    dbChirp.UserID,
	}
}

func getCleanedChirpBody(chirpBody string) string {
	var cleanedChirpBody string

	bannedWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	for _, word := range strings.Fields(chirpBody) {
		if _, ok := bannedWords[strings.ToLower(word)]; ok {
			cleanedChirpBody = cleanedChirpBody + " ****"
			continue
		}
		cleanedChirpBody = cleanedChirpBody + " " + word
	}

	return strings.TrimSpace(cleanedChirpBody)
}
