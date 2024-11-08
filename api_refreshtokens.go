package main

import (
	"context"
	"net/http"
	"time"

	"github.com/KidMuon/chirpy/internal/auth"
	"github.com/KidMuon/chirpy/internal/database"
)

func makeRefreshToken(cfg *apiConfig, user User) (string, error) {
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		return "", err
	}

	refreshTokenToCreate := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.Id,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	}

	_, err = cfg.db.CreateRefreshToken(context.Background(), refreshTokenToCreate)
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 400, "no refresh token present")
		return
	}

	dbRefreshToken, err := cfg.db.FindRefreshToken(context.Background(), refreshToken)
	if err != nil || dbRefreshToken.Token == "" {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	// make a new access pin for the user
	token, err := auth.MakeJWT(dbRefreshToken.UserID, cfg.tokenSecret, time.Duration(3600*1e9))
	if err != nil {
		respondWithError(w, 500, "something went wrong")
		return
	}

	respondWithJSON(w, 200, AccessTokenFromRefresh{Token: token})
}

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 400, "no refresh token present")
		return
	}

	_, err = cfg.db.RevokeRefreshToken(context.Background(), refreshToken)
	if err != nil {
		respondWithError(w, 500, "something went wrong")
		return
	}

	respondWithJSON(w, 204, nil)
}

type AccessTokenFromRefresh struct {
	Token string `json:"token"`
}
