package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zic20/chirpy/internal/auth"
	"github.com/zic20/chirpy/internal/database"
)

type authParams struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

type ResponseUser struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {

	reqBody := authParams{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqBody)
	if err != nil {
		log.Printf("error parsing request body: %s\n", err.Error())
		respondWithError(w, http.StatusInternalServerError, "could not parse request body")
		return
	}

	hash, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		log.Printf("error hashing user password: %s", err.Error())
		respondWithError(w, http.StatusInternalServerError, "Error creating user acccount please try again later.")
		return
	}

	user, err := cfg.Db.CreateUser(r.Context(), database.CreateUserParams{
		Email: reqBody.Email, HashedPassword: hash,
	})
	if err != nil {
		log.Printf("error creating user: %s", err.Error())
		respondWithError(w, 400, "Something went wrong please try again later")
		return
	}

	resBody := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, 201, resBody)

}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	reqBody := authParams{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqBody)
	if err != nil {
		log.Printf("error parsing request body: %s", err.Error())
		respondWithError(w, http.StatusInternalServerError, "server could not parse request body")
		return
	}

	user, err := cfg.Db.GetUserByEmail(r.Context(), reqBody.Email)
	if err != nil {
		log.Printf("user not found: %s", err.Error())
		respondWithError(w, http.StatusUnauthorized, "username or password incorrect")
		return
	}

	match, err := auth.CheckPasswordHash(reqBody.Password, user.HashedPassword)
	if err != nil {
		log.Printf("could not verify password: %s", err.Error())
		respondWithError(w, http.StatusInternalServerError, "Something went wrong please try again.")
		return
	}

	if !match {
		log.Print("incorrect password")
		respondWithError(w, http.StatusUnauthorized, "username or password incorrect")
		return
	}
	expires_in := time.Hour
	if reqBody.ExpiresInSeconds > 0 && reqBody.ExpiresInSeconds < 3600 {
		expires_in = time.Second * time.Duration(reqBody.ExpiresInSeconds)
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwt_secret, expires_in)

	if err != nil {
		log.Printf("Error forming jwt: %s", err.Error())
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, ResponseUser{ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	})
}
