package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/KidMuon/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type requestUser struct {
		Email string `json:"email"`
	}

	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	var reqUser requestUser
	err := decoder.Decode(&reqUser)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
	}

	userToCreate := database.CreateUserParams{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     reqUser.Email,
	}
	dbUser, err := cfg.db.CreateUser(r.Context(), userToCreate)
	if err != nil {
		respondWithError(w, 400, "email already in use")
		return
	}

	respondWithJSON(w, 201, dbUserToUser(dbUser))

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
