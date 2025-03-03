package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"unicode/utf8"

	"github.com/k3vwdd/chirpyWS/internal/types"
	"github.com/k3vwdd/chirpyWS/internal/utils"
)

type ApiConfig struct {
	types.ApiConfig
	FileServerHits atomic.Int32
}

func HandleChirpChars(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusNotFound)
		return
	}

	defer r.Body.Close()

	type requestBody struct {
		Body string `json:"body"`
	}

	type responseBody struct {
		//Valid string `json:"valid"`
		CleanedBody string `json:"cleaned_body"`
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		utils.RespondWithErrorHelper(w, 500, "couldn't read request")
		return
	}

	params := requestBody{}
	err = json.Unmarshal(data, &params)
	if err != nil {
		utils.RespondWithErrorHelper(w, 500, "couldn't unmarshal parameters")
		return
	}

	chirpCount := utf8.RuneCountInString(params.Body)
	if chirpCount > 140 {
		utils.RespondWithErrorHelper(w, 400, "Chirp is too long")
		return
	}

	cleanedWords := utils.CheckBadChirpLang(params.Body)

	utils.RespondWithJSONHelper(w, 200, responseBody{
		//Valid: "true",
		CleanedBody: cleanedWords,
	})
}

func (cfg *ApiConfig) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	cfg.FileServerHits.Store(0) // Resets the Hits.. their version
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	//fmt.Fprintf(w, "Hits: %v", cfg.fileServerHits.Sub(cfg.fileServerHits.Load()))  Resets the Hits.. my version
}

func HandleHealthReadiness(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

// Create a new handler that writes the number of requests that have been counted as plain text in this format to the HTTP response:
// Hits: x

func (cfg *ApiConfig) HandleWriteHits(w http.ResponseWriter, r *http.Request) {
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, html, cfg.FileServerHits.Load())
}
