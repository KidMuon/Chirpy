package main

import (
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

	respondWithJSON(w, 201, dbUserToUser(dbUser))
}

func (cfg *apiConfig) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	reqUser, resErr := getUserFromRequest(r)
	if resErr.err != nil {
		respondWithError(w, resErr.code, resErr.Error())
	}

	dbUser, err := cfg.db.GetUserByEmail(r.Context(), reqUser.Email)
	if err != nil {
		respondWithError(w, 401, "Incorrect Email or Password")
		return
	}

	err = auth.CheckPasswordHash(reqUser.Password, dbUser.HashedPassword)
	if err != nil {
		respondWithError(w, 401, "Incorrect Email or Password")
		return
	}

	respondWithJSON(w, 200, dbUserToUser(dbUser))
}

func getUserFromRequest(r *http.Request) (requestUser, responseError) {
	defer r.Body.Close()

	var reqUser requestUser
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqUser)
	if err != nil {
		return requestUser{}, responseError{code: 500, err: fmt.Errorf("Something went wrong")}
	}

	if reqUser.Password == "" {
		return requestUser{}, responseError{code: 400, err: fmt.Errorf("Password Required")}
	}

	hashed_password, err := auth.HashPassword(reqUser.Password)
	if err != nil {
		return requestUser{}, responseError{code: 500, err: fmt.Errorf("Something went wrong")}
	}
	reqUser.hashed_password = hashed_password

	return reqUser, responseError{}
}

type requestUser struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	hashed_password string
}

type User struct {
	Id         uuid.UUID `json:"id"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
	Email      string    `json:"email"`
}

func dbUserToUser(dbUser database.User) User {
	return User{
		Id:         dbUser.ID,
		Created_at: dbUser.CreatedAt,
		Updated_at: dbUser.UpdatedAt,
		Email:      dbUser.Email,
	}
}
