package main

import (
	"log"
	"net/http"
)

func main() {
	handler := http.NewServeMux()

	handler.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	handler.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	log.Printf("Server started at %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
