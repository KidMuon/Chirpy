package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/KidMuon/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverhits atomic.Int32
	db             *database.Queries
}

func main() {
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("Cannot connect to database")
	}
	dbQueries := database.New(db)

	var cfg apiConfig
	cfg.db = dbQueries

	mux := http.NewServeMux()
	appHandler := http.FileServer(http.Dir("."))
	mux.Handle("/app/", http.StripPrefix("/app", cfg.middlewareMetricsInc(appHandler)))
	mux.HandleFunc("GET /api/healthz", handleHealthz)
	mux.HandleFunc("GET /admin/metrics", cfg.handleServeMetric)
	mux.HandleFunc("POST /admin/reset", cfg.handleResetMetric)
	mux.HandleFunc("POST /api/validate_chirp", cfg.handleValidateChirp)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverhits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleServeMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("<html><body> <h1>Welcome, Chirpy Admin</h1> <p>Chirpy has been visited %d times!</p> </body> </html>", cfg.fileserverhits.Load())))
}

func (cfg *apiConfig) handleResetMetric(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverhits.Store(0)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("Count Reset")))
}

func (cfg *apiConfig) handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	type Chirp struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	var newChirp Chirp
	err := decoder.Decode(&newChirp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"error\":\"Something went wrong\"}"))
		return
	}

	var chirpCharacterCount int
	for range newChirp.Body {
		chirpCharacterCount++
	}

	if chirpCharacterCount > 140 {
		w.WriteHeader(400)
		w.Write([]byte("{\"error\":\"Chirp is too long\"}"))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("{\"cleaned_body\":\"%s\"}", getCleanedChirpBody(newChirp.Body))))
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
