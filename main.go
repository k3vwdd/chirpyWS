package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/k3vwdd/chirpyWS/internal/database"
	"github.com/k3vwdd/chirpyWS/internal/handlers"
	"github.com/k3vwdd/chirpyWS/internal/middleWare"
	"github.com/k3vwdd/chirpyWS/internal/types"
	_ "github.com/lib/pq"
)

func main() {
    godotenv.Load()
    jwtKey := os.Getenv("JWTKEY")
    dbURL := os.Getenv("DB_URL")
    dbDevURL := os.Getenv("PLATFORM")
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatalf("Error opening database: %v", err)
    }

    dbQueries := database.New(db)

    apiCfg := &types.ApiConfig{
        Db: dbQueries,
        Platform: dbDevURL,
        ApiKey: jwtKey,
    }

	cfg := &handlers.ApiConfig{
        ApiConfig: apiCfg,
    }

	mw := &middleWare.ApiConfig{
        ApiConfig: apiCfg,
    }

	filepathRoot := "."
	mux := http.NewServeMux()
	loggedMux := middleWare.MiddlewareLogging(mux)
	mux.Handle("/", http.FileServer(http.Dir(filepathRoot)))
	// strips "/" off of /app/

	mux.Handle("/app/", http.StripPrefix("/app/", mw.MiddlewareMetricsInc((http.FileServer(http.Dir(filepathRoot))))))
	mux.HandleFunc("GET /api/healthz", cfg.HandleHealthReadiness)
    mux.HandleFunc("GET /api/chirps", cfg.HandleGetChirps)
    mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.HandleGetSingleChirp)
    mux.HandleFunc("GET /admin/metrics", cfg.HandleWriteHits)
    mux.HandleFunc("POST /api/users", cfg.HandleCreateUser)
    mux.HandleFunc("POST /api/chirps", cfg.HandleCreateChirp)
    mux.HandleFunc("POST /api/login", cfg.HandleLogin)
    mux.HandleFunc("POST /api/refresh", cfg.HandleRefresh)
    mux.HandleFunc("POST /api/revoke", cfg.HandleRevokeToken)
	mux.HandleFunc("POST /admin/reset", cfg.HandleRegister)
    mux.HandleFunc("PUT /api/users", cfg.HandleUpdateUser)

	port := "8080"
	// a struct that describes a server configuration
	server := &http.Server{
		Addr:    ":" + port,
		Handler: loggedMux,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
}
