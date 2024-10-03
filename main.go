package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverhits atomic.Int32
}

func main() {
	var cfg apiConfig

	mux := http.NewServeMux()
	appHandler := http.FileServer(http.Dir("."))
	mux.Handle("/app/", http.StripPrefix("/app", cfg.middlewareMetricsInc(appHandler)))
	mux.HandleFunc("/healthz/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/metrics/", cfg.handleServeMetric)
	mux.HandleFunc("/reset/", cfg.handleResetServeMetric)
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverhits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleServeMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverhits.Load())))
}

func (cfg *apiConfig) handleResetServeMetric(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverhits.Store(0)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("Count Reset")))
}
