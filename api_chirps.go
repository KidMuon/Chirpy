package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"strings"

	"github.com/KidMuon/chirpy/internal/auth"
	"github.com/KidMuon/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handleGetAllChirps(w http.ResponseWriter, r *http.Request) {
	var authorIDPresent bool

	authorString := r.URL.Query().Get("author_id")

	authorID, err := uuid.Parse(authorString)
	if err != nil {
		respondWithError(w, 500, "something went wrong")
		return
	} else if authorString == "" {
		authorIDPresent = false
	} else {
		authorIDPresent = true
	}

	dbChirps := []database.Chirp{}

	if !authorIDPresent {
		allDBChirps, err := cfg.db.GetAllChirps(r.Context())
		if err != nil {
			respondWithError(w, 500, "something went wrong")
			return
		}
		dbChirps = append(dbChirps, allDBChirps...)
	} else {
		authorDBChirps, err := cfg.db.GetAllChirpsByAuthor(context.Background(), authorID)
		if err != nil {
			respondWithError(w, 500, "something went wrong")
			return
		}
		dbChirps = append(dbChirps, authorDBChirps...)
	}

	chirps := []Chirp{}
	for _, dbChrip := range dbChirps {
		chirps = append(chirps, dbChirpToChirp(dbChrip))
	}

	respondWithJSON(w, 200, chirps)
}

func (cfg *apiConfig) handleGetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpUUID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 500, "something went wrong")
		return
	}
	dbChirp, err := cfg.db.GetChirpByID(r.Context(), chirpUUID)
	if err != nil {
		respondWithError(w, 404, "not found")
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
		respondWithError(w, 500, "something went wrong")
		return
	}

	var chirpCharacterCount int
	for range reqChirp.Body {
		chirpCharacterCount++
	}

	if chirpCharacterCount > 140 {
		respondWithError(w, 400, "chirp is too long")
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

func (cfg *apiConfig) handleDeleteChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpUUID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 500, "something went wrong")
		return
	}

	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "no authentication found")
		return
	}
	userID, err := auth.ValidateJWT(authToken, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	dbChirp, err := cfg.db.GetChirpByID(r.Context(), chirpUUID)
	if err != nil {
		respondWithError(w, 404, "not found")
		return
	}
	if dbChirp.UserID != userID {
		respondWithError(w, 403, "unauthorized")
		return
	}

	deletedDbChirp, err := cfg.db.DeleteChirpByID(context.Background(), chirpUUID)
	if err != nil {
		respondWithError(w, 500, "something went wrong")
		return
	}

	respondWithJSON(w, 204, dbChirpToChirp(deletedDbChirp))
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
