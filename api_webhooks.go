package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/KidMuon/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handleWebhooks(w http.ResponseWriter, r *http.Request) {
	reqApiKey, err := auth.GetPolkaAPIKey(r.Header)
	if err != nil {
		respondWithError(w, 401, "authentication error")
		return
	}

	if reqApiKey != cfg.polkaKey {
		respondWithError(w, 401, "unauthorized")
		return
	}

	type webhookRequest struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	defer r.Body.Close()
	var reqWebHook webhookRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&reqWebHook)
	if err != nil {
		respondWithError(w, 400, "malformed request")
		return
	}

	switch reqWebHook.Event {
	case "user.upgraded":
		handleMakeUserChirpyRed(w, cfg, reqWebHook.Data.UserID)
	default:
		respondWithJSON(w, 204, User{})
	}
}

func handleMakeUserChirpyRed(w http.ResponseWriter, cfg *apiConfig, userID uuid.UUID) {
	dbUser, err := cfg.db.AddChirpyRedByID(context.Background(), userID)
	if err != nil {
		respondWithError(w, 404, "user not found")
		return
	}
	respondWithJSON(w, 204, dbUserToUser(dbUser))
}
