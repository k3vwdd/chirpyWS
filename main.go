package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"unicode/utf8"

	"go.uber.org/atomic"
)

type apiConfig struct {
    fileServerHits atomic.Int32
}
func middlewareLogging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        bodyBytes, err := io.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "Failed to read request body", http.StatusInternalServerError)
            return
        }
        log.Printf("Request body: %s\n", string(bodyBytes))
        // Put it back so the handler can read it again
        r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
        next.ServeHTTP(w, r)
    })
}


func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cfg.fileServerHits.Add(1)
        next.ServeHTTP(w, r)
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

func respondWithJSONHelper(w http.ResponseWriter, statuscode int, payload interface{}) error {
    response, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.WriteHeader(statuscode)
    w.Write(response)
    return nil
}

func respondWithErrorHelper(w http.ResponseWriter, statuscode int, msg string) error {
    return respondWithJSONHelper(w, statuscode, map[string]string{"error": msg})
}

func handleChirpChars(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method Not Allowed", http.StatusNotFound)
        return
    }

    defer r.Body.Close()

    type requestBody struct {
        Body string `json:"body"`
    }

    type responseBody struct {
        Valid string `json:"valid"`
    }

    data, err := io.ReadAll(r.Body)
    if err != nil {
        respondWithErrorHelper(w, 500, "couldn't read request")
        return
    }

    params := requestBody{}
    err = json.Unmarshal(data, &params)
    if err != nil {
        respondWithErrorHelper(w, 500, "couldn't unmarshal parameters")
        return
    }

    characterCount := utf8.RuneCountInString(params.Body)
    if characterCount > 140 {
        respondWithErrorHelper(w, 400, "Chirp is too long")
        return
    }

    respondWithJSONHelper(w, 200, responseBody{
        Valid: "true",
    })
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
    loggedMux := middlewareLogging(mux)
    mux.Handle("/", http.FileServer(http.Dir(filepathRoot)))
    // strips "/" off of /app/

    mux.Handle("/app/", http.StripPrefix("/app/", cfg.middlewareMetricsInc((http.FileServer(http.Dir(filepathRoot))))))
    mux.HandleFunc("/api/healthz", handleHealthReadiness)
    mux.HandleFunc("/api/validate_chirp", handleChirpChars)
    mux.HandleFunc("/admin/metrics", cfg.handleWriteHits)
    mux.HandleFunc("/admin/reset", cfg.handleRegister)

    port := "8080"
    // a struct that describes a server configuration
    server := &http.Server{
        Addr: ":" + port,
        Handler: loggedMux,
    }

    err := server.ListenAndServe()
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
}
