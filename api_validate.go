package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (cfg *apiConfig) handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type Chirp struct {
		Body string `json:"body"`
	}

	type ValidatedChirp struct {
		Cleaned_Body string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	var newChirp Chirp
	err := decoder.Decode(&newChirp)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	var chirpCharacterCount int
	for range newChirp.Body {
		chirpCharacterCount++
	}

	if chirpCharacterCount > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	respondWithJSON(w, 200, ValidatedChirp{Cleaned_Body: getCleanedChirpBody(newChirp.Body)})
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
