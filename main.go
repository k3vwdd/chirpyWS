package main

import (
	"log"
	"net/http"
	"github.com/k3vwdd/chirpyWS/internal/handlers"
	"github.com/k3vwdd/chirpyWS/internal/middleWare"
)

func main() {
	cfg := &handlers.ApiConfig{}
	mw := &middleWare.ApiConfig{}

	filepathRoot := "."
	mux := http.NewServeMux()
	loggedMux := middleWare.MiddlewareLogging(mux)
	mux.Handle("/", http.FileServer(http.Dir(filepathRoot)))
	// strips "/" off of /app/

	mux.Handle("/app/", http.StripPrefix("/app/", mw.MiddlewareMetricsInc((http.FileServer(http.Dir(filepathRoot))))))
	mux.HandleFunc("/api/healthz", handlers.HandleHealthReadiness)
	mux.HandleFunc("/api/validate_chirp", handlers.HandleChirpChars)
	mux.HandleFunc("/admin/metrics", cfg.HandleWriteHits)
	mux.HandleFunc("/admin/reset", cfg.HandleRegister)

	port := "8080"
	// a struct that describes a server configuration
	server := &http.Server{
		Addr:    ":" + port,
		Handler: loggedMux,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
}
