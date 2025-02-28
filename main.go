package main

import (
	"fmt"
	"log"
	"net/http"
)


func main() {

    mux := http.NewServeMux()
    mux.Handle("/", http.FileServer(http.Dir(".")))

    server := &http.Server{
        Addr: ":8080",
        Handler: mux,
    }

    err := server.ListenAndServe()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Server started")
}
