package main

import (
	"log"
	"net/http"
)

func main() {
	handler := http.NewServeMux()

	handler.Handle("/", http.FileServer(http.Dir(".")))

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	log.Printf("Server started at %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
