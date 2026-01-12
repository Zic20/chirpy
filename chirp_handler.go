package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zic20/chirpy/internal/auth"
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
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, http.StatusUnauthorized, err.Error())
	}
	userid, err := auth.ValidateJWT(token, cfg.jwt_secret)
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, http.StatusUnauthorized, err.Error())
	}
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

	err = decoder.Decode(&reqBody)
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
		UserID: userid,
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

func (cfg *apiConfig) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpid := r.PathValue("chirpID")
	chirp, err := cfg.Db.GetChirp(r.Context(), uuid.MustParse(chirpid))
	if err != nil {
		log.Printf("could not fetch chirp from the database: %s", err.Error())
		respondWithError(w, http.StatusNotFound, "Not found")
		return
	}

	respondWithJSON(w, http.StatusOK, ChirpResponse{
		chirp.ID, chirp.CreatedAt, chirp.UpdatedAt, chirp.Body, chirp.UserID,
	})
}

func (cfg *apiConfig) handleGetAllChirps(w http.ResponseWriter, r *http.Request) {
	response := []ChirpResponse{}

	chirps, err := cfg.Db.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("Error fetching chirps: %s", err.Error())
		respondWithError(w, http.StatusInternalServerError, "error fetching chirps")
		return
	}

	for _, chirp := range chirps {
		response = append(response, ChirpResponse{chirp.ID, chirp.CreatedAt, chirp.UpdatedAt, chirp.Body, chirp.UserID})
	}

	respondWithJSON(w, http.StatusOK, response)
}
