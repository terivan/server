package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func ReadinessHandlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) VaridateChirpHandlerFunc(w http.ResponseWriter, r *http.Request) {

	type Chirp struct {
		Body string `json:"body"`
	}

	type LengthError struct {
		Error string `json:"error"`
	}

	type ValidReturn struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)

	var NewChirp Chirp
	err := decoder.Decode(&NewChirp)

	if err != nil {
		log.Printf("Failed decoding chirp, %v", err)
		w.WriteHeader(500)
	}

	stringLength := len(NewChirp.Body)
	w.Header().Set("Content-Type", "application/json")

	if stringLength > 140 {
		responseBody := LengthError{
			Error: "Chirp is too long",
		}

		dat, err := json.Marshal(responseBody)

		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(400)
		w.Write(dat)
		return
	}

	validResponseBody := ValidReturn{
		Valid: true,
	}

	valid, err := json.Marshal(validResponseBody)

	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	w.Write(valid)
	return

}

func (cfg *apiConfig) MetricsHandlerFunc(w http.ResponseWriter, r *http.Request) {
	count := cfg.fileserverHits.Load()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	response := fmt.Sprintf(`<html>
							  <body>
							    <h1>Welcome, Chirpy Admin</h1>
							    <p>Chirpy has been visited %d times!</p>
							  </body>
							</html>`, count)
	w.Write([]byte(response))
}

func (cfg *apiConfig) ResetMetricsHandlerFunc(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
}

func main() {
	mux := http.NewServeMux()

	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	cfg := apiConfig{}

	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer)))
	mux.HandleFunc("GET /api/healthz", ReadinessHandlerFunc)
	mux.HandleFunc("GET /admin/metrics", cfg.MetricsHandlerFunc)
	mux.HandleFunc("POST /admin/reset", cfg.ResetMetricsHandlerFunc)
	mux.HandleFunc("POST /api/validate_chirp", cfg.VaridateChirpHandlerFunc)

	err := server.ListenAndServe()

	if err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}

	return
}
