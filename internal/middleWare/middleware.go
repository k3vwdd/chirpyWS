package middleWare

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/k3vwdd/chirpyWS/internal/types"
)

type ApiConfig struct {
	types.ApiConfig
}

func MiddlewareLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// read the data while simultaneously writing it to a buffer for logging.
		// This allows you to pass the original stream to the next handler without needing to reassemble it completely.
		var buff bytes.Buffer
		tee := io.TeeReader(r.Body, &buff)
		r.Body = io.NopCloser(tee)
		next.ServeHTTP(w, r)
		log.Printf("Request body: %s\n", buff.String())
		//bodyBytes, err := io.ReadAll(r.Body)
		//if err != nil {
		//    http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		//    return
		//}
		//log.Printf("Request body: %s\n", string(bodyBytes))
		//// reset the body and Put it back so the handler can read it again. 'No operation' a nil 'Close'
		//// stuff the request body back with the input
		//r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		//next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
