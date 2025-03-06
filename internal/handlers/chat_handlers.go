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
	"github.com/k3vwdd/chirpyWS/internal/auth"
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
        Password string `json:"password"`
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

    hashPassword, err := auth.HashPassword(params.Password)
    if err != nil {
		utils.RespondWithErrorHelper(w, 500, "Unable to hash password")
		return
    }

    user, err := cfg.Db.CreateUser(r.Context(), database.CreateUserParams{
        ID: uuid.New(),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Email: params.Email,
        HashedPassword: hashPassword,
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
		utils.RespondWithErrorHelper(w, 404, "Unauthorized: Method Not Allowed")
        return
    }

    defer r.Body.Close()

	type requestBody struct {
		Body string `json:"body"`
	}

	type responseBody struct {
		Id uuid.UUID `json:"id"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`
		Body string `json:"body"`
        UserId uuid.UUID `json:"user_id"`
	}

    authHeader := r.Header
    tokenString, err := auth.GetBearerToken(authHeader)
    if err != nil {
		utils.RespondWithErrorHelper(w, 404, "Unauthorized: Invalid header")
        return
    }

    userID, err := auth.ValidateJWT(tokenString, cfg.ApiKey)
    if err != nil {
		utils.RespondWithErrorHelper(w, 401, "Unauthorized: Invalid token")
        return
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
        UserID: userID,
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
        UserId: userID,
	})
}

func (cfg *ApiConfig) HandleGetChirps(w http.ResponseWriter, r *http.Request) {

    if r.Method != http.MethodGet {
        utils.RespondWithErrorHelper(w, 400, "method not allowed")
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

func (cfg *ApiConfig) HandleGetSingleChirp(w http.ResponseWriter, r *http.Request) {

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

    requestedChirp := r.PathValue("chirpID")
    parsedChirp, err := uuid.Parse(requestedChirp)
    if err != nil {
        utils.RespondWithErrorHelper(w, 500, "unable to convert string to uuid")
        return
    }

    getChirp, err := cfg.Db.GetChirpByID(r.Context(), parsedChirp)
    if err != nil {
        utils.RespondWithErrorHelper(w, 400, "chirpID doesn't exists")
        return
    }

    chirp := responseBody{
            Id: getChirp.ID,
            CreatedAt: getChirp.CreatedAt,
            UpdatedAt: getChirp.UpdatedAt,
            Body: getChirp.Body,
            UserId: getChirp.UserID,
        }

        utils.RespondWithJSONHelper(w, 200, chirp)
    }

func (cfg *ApiConfig) HandleLogin(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method Not Allowed", http.StatusNotFound)
        return
    }

    defer r.Body.Close()

	type requestBody struct {
		Email string `json:"email"`
        Password string `json:"password"`
        //ExpiresInSeconds *int `json:"expires_in_seconds"`
	}

    type responseBody struct {
        Id uuid.UUID `json:"id"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`
        Email string `json:"email"`
        Token string `json:"token"`
        RefreshToken string `json:"refresh_token"`
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

    getUser, err := cfg.Db.GetUserByEmail(r.Context(), params.Email)
    if err != nil {
		utils.RespondWithErrorHelper(w, 400, "Incorrect email or password")
		return
    }

    checkPass := auth.CheckPasswordHash(params.Password, getUser.HashedPassword)
    if checkPass != nil {
		utils.RespondWithErrorHelper(w, 400, "Incorrect email or password")
        return
    }

    refreshtoken, err := auth.MakeRefreshToken()
    if err != nil {
        utils.RespondWithErrorHelper(w, 500, "Failed to generate refresh token")
        return
    }

    createRefreshToken := cfg.Db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
        Token: refreshtoken,
        UserID: getUser.ID,
    })

    if createRefreshToken != nil {
        utils.RespondWithErrorHelper(w, 500, "Failed to create refresh token in db")
    }

    //if params.ExpiresInSeconds != nil {
    //    requestedDuration := time.Duration(*params.ExpiresInSeconds) * time.Second

    //    if requestedDuration < defaultDuration {
    //        defaultDuration = requestedDuration
    //    }
    //}

    jwtToken, err := auth.MakeJWT(getUser.ID, cfg.ApiKey, time.Hour * 1)
    if err != nil {
        utils.RespondWithErrorHelper(w, 500, "Failed to generate JWT")
        return
    }

    user := responseBody{
        Id: getUser.ID,
        CreatedAt: getUser.CreatedAt,
        UpdatedAt: getUser.UpdatedAt,
        Email: getUser.Email,
        Token: jwtToken,
        RefreshToken: refreshtoken,
    }

	utils.RespondWithJSONHelper(w, 200, user)
}


func (cfg *ApiConfig) HandleRefresh(w http.ResponseWriter, r *http.Request) {
    type responseBody struct {
        Token string `json:"token"`
    }

    if r.Method != http.MethodPost {
        http.Error(w, "Method Not Allowed", http.StatusNotFound)
        return
    }

    authHeader := r.Header
    refreshTokenString, err := auth.GetBearerToken(authHeader)
    if err != nil {
		utils.RespondWithErrorHelper(w, 401, "Unauthorized: Invalid header")
        return
    }

    user, err := cfg.Db.GetUserFromRefreshToken(r.Context(), refreshTokenString)
    if err != nil {
		utils.RespondWithErrorHelper(w, 401, "Unable to retrieve token from user")
        return
    }

    jwtToken, err := auth.MakeJWT(user.ID, cfg.ApiKey, time.Hour * 1)
    if err != nil {
		utils.RespondWithErrorHelper(w, 401, "Error creating jwt with duration 1 hour")
        return
    }

    utils.RespondWithJSONHelper(w, 200, responseBody{
        Token: jwtToken,
    })

}

func (cfg *ApiConfig) HandleRevokeToken(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
        return
    }

    authHeader := r.Header
    refreshTokenString, err := auth.GetBearerToken(authHeader)
    if err != nil {
		utils.RespondWithErrorHelper(w, 401, "Unauthorized: Invalid header")
        return
    }

    err = cfg.Db.RevokeRefreshToken(r.Context(), refreshTokenString)
    if err != nil {
		utils.RespondWithErrorHelper(w, 401, "Unauthorized: Unable to revoke refresh token")
        return
    }

    utils.RespondWithJSONHelper(w, 204, "")

}

func (cfg *ApiConfig) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	type requestBody struct {
        Password string `json:"password"`
		Email string `json:"email"`
	}

	type responseBody struct {
		Id uuid.UUID `json:"id"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`
        Email string `json:"email"`
	}

    authHeader := r.Header
    tokenString, err := auth.GetBearerToken(authHeader)
    if err != nil {
		utils.RespondWithErrorHelper(w, 404, "Unauthorized: Invalid header")
        return
    }

    userID, err := auth.ValidateJWT(tokenString, cfg.ApiKey)
    if err != nil {
		utils.RespondWithErrorHelper(w, 401, "Unauthorized: Invalid token")
        return
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

    hashPassword, err := auth.HashPassword(params.Password)
    if err != nil {
		utils.RespondWithErrorHelper(w, 500, "Unable to hash password")
		return
    }

    err = cfg.Db.UpdateUserEmailAndPassword(r.Context(), database.UpdateUserEmailAndPasswordParams{
        Email: params.Email,
        ID: userID,
        HashedPassword: hashPassword,
    })

    if err != nil {
		utils.RespondWithErrorHelper(w, 500, "Error Updating email and password")
        return
    }

    user, err := cfg.Db.GetUserByID(r.Context(), userID)
    if err != nil {
		utils.RespondWithErrorHelper(w, 500, "Unalbe to fetch updated User")
        return
    }

	utils.RespondWithJSONHelper(w, 200, responseBody{
        Email: params.Email,
        Id: user.ID,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
	})
}
