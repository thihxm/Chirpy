package main

import "net/http"

func main() {
	handler := http.NewServeMux()

	handler.Handle("/", http.FileServer(http.Dir(".")))

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}
	server.ListenAndServe()
}
