package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

const (
	MaxChirpLength = 140
)

var profaneWords = map[string]bool{
	"kerfuffle": true,
	"sharbert":  true,
	"fornax":    true,
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	mux := http.NewServeMux()
	apiCfg := &apiConfig{}

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("POST /admin/reset", apiCfg.reset)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metrics)
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.Handle("/api/validate_chirp", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Body string `json:"body"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			log.Printf("Error decoding parameters: %v", err)
			respondWithError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		if len(params.Body) > MaxChirpLength {
			respondWithError(w, http.StatusBadRequest, "Chirp is too long")
			return
		}

		type returnVals struct {
			// the key will be the name of struct field unless you give it an explicit JSON tag
			CleanedBody string `json:"cleaned_body"`
		}

		resBody := returnVals{
			CleanedBody: removeProfanity(params.Body),
		}
		respondWithJSON(w, http.StatusOK, resBody)
	}))

	log.Printf("Server started at %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(
		`<html>
        <body>
            <h1>Welcome, Chirpy Admin</h1>
            <p>Chirpy has been visited %d times!</p>
        </body>
        </html>`,
		cfg.fileserverHits.Load())),
	)
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	w.Write([]byte(msg))
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling response JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(data)
}

func removeProfanity(body string) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		lowerCaseWord := strings.ToLower(word)
		if profaneWords[lowerCaseWord] {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}
