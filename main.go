package main

import (
	//"fmt"
	"log"
	"net/http"
)

func readinessHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        w.Header().Set("Content-Type" , "text/plain; charset=utf-8")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(http.StatusText(http.StatusOK)))
    }
}

func main() {

    filepathRoot := "."
    mux := http.NewServeMux()
    // strips "/" off of /app/ - remember you have /app/assets ....
    mux.Handle("/", http.FileServer(http.Dir(filepathRoot)))
    mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot))))
    mux.HandleFunc("/healthz", readinessHandler)

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
