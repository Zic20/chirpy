package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zic20/chirpy/internal/database"
)

type ChirpResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	type params struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	type errResBody struct {
		Error string `json:"error"`
	}
	type returnVal struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	reqBody := params{}

	err := decoder.Decode(&reqBody)
	if err != nil {
		log.Printf("Error parsing request body: %s\n", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if len(reqBody.Body) > 140 {
		log.Printf("Chirp was too long: %d characters\n", len(reqBody.Body))
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	blockedWords := map[string]struct{}{
		"herfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	text := cleanBody(reqBody.Body, blockedWords)

	chirp, err := cfg.Db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   text,
		UserID: reqBody.UserID,
	})

	if err != nil {
		log.Printf("error creating chirp: %s", err)
		respondWithError(w, http.StatusInternalServerError, "couldn't create chirp")
		return
	}
	responsse := ChirpResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	respondWithJSON(w, 201, responsse)
}

func cleanBody(body string, blockedWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		if _, ok := blockedWords[strings.ToLower(word)]; ok {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}
