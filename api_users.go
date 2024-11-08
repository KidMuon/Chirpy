package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/KidMuon/chirpy/internal/auth"
	"github.com/KidMuon/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	reqUser, resErr := getUserFromRequest(r)
	if resErr.err != nil {
		respondWithError(w, resErr.code, resErr.Error())
		return
	}

	userToCreate := database.CreateUserParams{
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Email:          reqUser.Email,
		HashedPassword: reqUser.hashed_password,
	}
	dbUser, err := cfg.db.CreateUser(r.Context(), userToCreate)
	if err != nil {
		respondWithError(w, 400, "email already in use")
		return
	}
	user := dbUserToUser(dbUser)

	token, err := auth.MakeJWT(user.Id, cfg.tokenSecret, reqUser.expiration_duration)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	user.Token = token

	respondWithJSON(w, 201, user)
}

func (cfg *apiConfig) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	reqUser, resErr := getUserFromRequest(r)
	if resErr.err != nil {
		respondWithError(w, resErr.code, resErr.Error())
		return
	}

	dbUser, err := cfg.db.GetUserByEmail(r.Context(), reqUser.Email)
	if err != nil {
		respondWithError(w, 401, "incorrect email or password")
		return
	}

	err = auth.CheckPasswordHash(reqUser.Password, dbUser.HashedPassword)
	if err != nil {
		respondWithError(w, 401, "incorrect email or password")
		return
	}
	user := dbUserToUser(dbUser)

	token, err := auth.MakeJWT(user.Id, cfg.tokenSecret, reqUser.expiration_duration)
	if err != nil {
		respondWithError(w, 500, "something went wrong")
		return
	}
	user.Token = token

	refreshToken, err := makeRefreshToken(cfg, user)
	if err != nil {
		respondWithError(w, 500, "something went wrong")
		return
	}
	user.RefreshToken = refreshToken

	respondWithJSON(w, 200, user)
}

func (cfg *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	reqUser, resErr := getUserFromRequest(r)
	if resErr.err != nil {
		respondWithError(w, resErr.code, resErr.Error())
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

	userEmail := reqUser.Email

	userHashedPassword, err := auth.HashPassword(reqUser.Password)
	if err != nil {
		respondWithError(w, 500, "something went wrong")
		return
	}

	userToUpdate := database.UpdateUserEmailAndPasswordParams{
		ID:             userID,
		Email:          userEmail,
		HashedPassword: userHashedPassword,
	}

	updatedDBUser, err := cfg.db.UpdateUserEmailAndPassword(context.Background(), userToUpdate)
	if err != nil {
		respondWithError(w, 500, "something went wrong")
		return
	}

	respondWithJSON(w, 200, dbUserToUser(updatedDBUser))
}

func getUserFromRequest(r *http.Request) (requestUser, responseError) {
	defer r.Body.Close()

	var reqUser requestUser
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqUser)
	if err != nil {
		return requestUser{}, responseError{code: 500, err: fmt.Errorf("something went wrong")}
	}

	if reqUser.Password == "" {
		return requestUser{}, responseError{code: 400, err: fmt.Errorf("password required")}
	}

	hashed_password, err := auth.HashPassword(reqUser.Password)
	if err != nil {
		return requestUser{}, responseError{code: 500, err: fmt.Errorf("something went wrong")}
	}
	reqUser.hashed_password = hashed_password

	reqUser.expiration_duration = time.Duration(3600 * 1e9)

	return reqUser, responseError{}
}

type requestUser struct {
	Email               string `json:"email"`
	Password            string `json:"password"`
	expiration_duration time.Duration
	hashed_password     string
}

type User struct {
	Id           uuid.UUID `json:"id"`
	Created_at   time.Time `json:"created_at"`
	Updated_at   time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func dbUserToUser(dbUser database.User) User {
	return User{
		Id:         dbUser.ID,
		Created_at: dbUser.CreatedAt,
		Updated_at: dbUser.UpdatedAt,
		Email:      dbUser.Email,
	}
}
