package main

import (
	"log"
	"net/http"
)

func ReadinessHandlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func main() {
	mux := http.NewServeMux()

	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/app/", http.StripPrefix("/app", fileServer))
	mux.HandleFunc("/healthz", ReadinessHandlerFunc)

	err := server.ListenAndServe()

	if err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}

	return
}
