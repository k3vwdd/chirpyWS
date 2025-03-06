package types

import (
    "sync/atomic"
	"github.com/k3vwdd/chirpyWS/internal/database"
)


type ApiConfig struct {
	FileServerHits atomic.Int32
    Db *database.Queries
    Platform string
    ApiKey  string
}
