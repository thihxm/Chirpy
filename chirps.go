package main

import (
	"net/http"

	"github.com/thihxm/Chirpy/internal/config"
)

func createChirpHandler(cfg *config.ApiConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}
