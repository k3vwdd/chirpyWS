package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/k3vwdd/chirpyWS/internal/database"
	"github.com/k3vwdd/chirpyWS/internal/types"
	"github.com/k3vwdd/chirpyWS/internal/utils"
)

type ApiConfig struct {
	*types.ApiConfig
	FileServerHits atomic.Int32
}


func (cfg *ApiConfig) HandleRegister(w http.ResponseWriter, r *http.Request) {
    if cfg.ApiConfig.Platform != "dev" {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

    err := cfg.Db.DeleteAllUsers(r.Context())
    if err != nil {
		utils.RespondWithErrorHelper(w, 400, "Unable to delete users from DB")
		return
    }

    utils.RespondWithJSONHelper(w, 200, "Success.. All users removed from db")

	//cfg.FileServerHits.Store(0) // Resets the Hits.. their version
	//w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	//w.WriteHeader(http.StatusOK)
	//fmt.Fprintf(w, "Hits: %v", cfg.fileServerHits.Sub(cfg.fileServerHits.Load()))  Resets the Hits.. my version
}

func (cfg *ApiConfig) HandleHealthReadiness(w http.ResponseWriter, r *http.Request) {
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

func (cfg *ApiConfig) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusNotFound)
		return
	}

	defer r.Body.Close()

	type requestBody struct {
		Email string `json:"email"`
	}

	type responseBody struct {
		Id uuid.UUID `json:"id"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`
        Email string `json:"email"`
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

    user, err := cfg.Db.CreateUser(r.Context(), database.CreateUserParams{
        ID: uuid.New(),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Email: params.Email,
    })

    if err != nil {
		fmt.Fprintf(os.Stderr, "Error Creating user: %v\n", err)
    }

	utils.RespondWithJSONHelper(w, 201, responseBody{
        Id: user.ID,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
        Email: user.Email,
	})
}

func (cfg *ApiConfig) HandleCreateChirp(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method Not Allowed", http.StatusNotFound)
        return
    }

    defer r.Body.Close()

	type requestBody struct {
		Body string `json:"body"`
        UserId uuid.UUID `json:"user_id"`
	}

	type responseBody struct {
		Id uuid.UUID `json:"id"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`
		Body string `json:"body"`
        UserId uuid.UUID `json:"user_id"`
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		utils.RespondWithErrorHelper(w, 500, "couldn't read request")
		return
	}

	params := requestBody{}
	err = json.Unmarshal(data, &params)
	if err != nil {
		utils.RespondWithErrorHelper(w, 500, "error with json format")
		return
	}

	chirpCount := utf8.RuneCountInString(params.Body)
	if chirpCount > 140 {
		utils.RespondWithErrorHelper(w, 400, "Chirp is too long")
		return
	}

    cleanedWords := utils.CheckBadChirpLang(params.Body)

    chirp, err := cfg.Db.CreateChirp(r.Context(), database.CreateChirpParams{
        ID:        uuid.New(),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Body:      cleanedWords,
        UserID:    params.UserId,
    })

    if err != nil {
		fmt.Fprintf(os.Stderr, "Error Creating chirp: %v\n", err)
		utils.RespondWithErrorHelper(w, 500, "Error creating chirp")
        return
    }

	utils.RespondWithJSONHelper(w, 201, responseBody{
        Id: chirp.ID,
        CreatedAt: chirp.CreatedAt,
        UpdatedAt: chirp.UpdatedAt,
        Body: cleanedWords,
        UserId: chirp.UserID,
	})
}

func (cfg *ApiConfig) HandleGetChirps(w http.ResponseWriter, r *http.Request) {

    if r.Method != http.MethodGet {
        http.Error(w, "Method Not Allowed", http.StatusNotFound)
        return
    }

	type responseBody struct {
		Id uuid.UUID `json:"id"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`
		Body string `json:"body"`
        UserId uuid.UUID `json:"user_id"`
	}

    getChirps, err := cfg.Db.GetAllChirps(r.Context())
    if err != nil {
        utils.RespondWithErrorHelper(w, 500, "error getting chirps from DB")
        return
    }

    result := []responseBody{}
    for _, val := range getChirps {
        chirps := []responseBody{
            {
                Id: val.ID,
                CreatedAt: val.CreatedAt,
                UpdatedAt: val.UpdatedAt,
                Body: val.Body,
                UserId: val.UserID,
            },
        }
        result = append(result, chirps...)
    }

    utils.RespondWithJSONHelper(w, 201, result)
}


