package main

import (
	"log"
	"net/http"
)

func main() {
	handler := http.NewServeMux()

	handler.Handle("/", http.FileServer(http.Dir(".")))
	handler.Handle("/assets", http.FileServer(http.Dir("./assets")))

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	log.Printf("Server started at %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
