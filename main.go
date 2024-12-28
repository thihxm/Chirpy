package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/thihxm/Chirpy/internal/config"
	"github.com/thihxm/Chirpy/internal/database"
	"github.com/thihxm/Chirpy/internal/utils"
)

const (
	MaxChirpLength = 140
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
		return
	}
	dbQueries := database.New(db)

	mux := http.NewServeMux()
	cfg := &config.ApiConfig{
		Queries: dbQueries,
	}

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("POST /admin/reset", cfg.Reset)
	mux.HandleFunc("GET /admin/metrics", cfg.Metrics)
	mux.Handle("/app/", cfg.MiddlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.Handle("POST /api/validate_chirp", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Body string `json:"body"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			log.Printf("Error decoding parameters: %v", err)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		if len(params.Body) > MaxChirpLength {
			utils.RespondWithError(w, http.StatusBadRequest, "Chirp is too long")
			return
		}

		type returnVals struct {
			// the key will be the name of struct field unless you give it an explicit JSON tag
			CleanedBody string `json:"cleaned_body"`
		}

		resBody := returnVals{
			CleanedBody: utils.RemoveProfanity(params.Body),
		}
		utils.RespondWithJSON(w, http.StatusOK, resBody)
	}))

	mux.Handle("POST /api/users", createUserHandler(cfg))

	log.Printf("Server started at %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
