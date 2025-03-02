package main

import (
	"fmt"
	"log"
	"net/http"

	"go.uber.org/atomic"
)

type apiConfig struct {
    fileServerHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        next.ServeHTTP(w, r)
        cfg.fileServerHits.Add(1)
    })
}

// Create a new handler that writes the number of requests that have been counted as plain text in this format to the HTTP response:
// Hits: x

func (cfg *apiConfig) handleWriteHits(w http.ResponseWriter, r *http.Request) {
    html := `
     <html>
      <body>
        <h1>Welcome, Chirpy Admin</h1>
        <p>Chirpy has been visited %v times!</p>
      </body>
    </html>
    `
    if r.Method != http.MethodGet {
        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
        return
    }
    w.Header().Set("Content-Type" , "text/html; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, html, cfg.fileServerHits.Load())
}

func (cfg *apiConfig) handleRegister(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
        return
    }
    cfg.fileServerHits.Store(0) // Resets the Hits.. their version
    w.Header().Set("Content-Type" , "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    //fmt.Fprintf(w, "Hits: %v", cfg.fileServerHits.Sub(cfg.fileServerHits.Load()))  Resets the Hits.. my version
}

func handleHealthReadiness(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
        return
    }
    w.Header().Set("Content-Type" , "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(http.StatusText(http.StatusOK)))
}

func main() {

    cfg := &apiConfig{}

    filepathRoot := "."
    mux := http.NewServeMux()
    mux.Handle("/", http.FileServer(http.Dir(filepathRoot)))
    // strips "/" off of /app/

    mux.Handle("/app/", http.StripPrefix("/app/", cfg.middlewareMetricsInc((http.FileServer(http.Dir(filepathRoot))))))
    mux.HandleFunc("/api/healthz", handleHealthReadiness)
    mux.HandleFunc("/admin/metrics", cfg.handleWriteHits)
    mux.HandleFunc("/admin/reset", cfg.handleRegister)

    port := "8080"
    // a struct that describes a server configuration
    server := &http.Server{
        Addr: ":" + port,
        Handler: mux,
    }

    err := server.ListenAndServe()
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
}
