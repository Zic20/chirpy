package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Email string `json:"email"`
	}

	reqBodyJson := params{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqBodyJson)
	if err != nil {
		log.Printf("error parsing request body: %s\n", err)
		respondWithError(w, 500, "server could not parse request body")
		return
	}
	user, err := cfg.Db.CreateUser(r.Context(), reqBodyJson.Email)
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
