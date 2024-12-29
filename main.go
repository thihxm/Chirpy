package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/thihxm/Chirpy/internal/config"
	"github.com/thihxm/Chirpy/internal/database"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	authSecret := os.Getenv("AUTH_SECRET")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
		return
	}
	dbQueries := database.New(db)

	mux := http.NewServeMux()
	cfg := &config.ApiConfig{
		Queries:    dbQueries,
		AuthSecret: authSecret,
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

	mux.Handle("POST /api/users", createUserHandler(cfg))
	mux.Handle("PUT /api/users", middlewareIsAuthenticated(cfg, updateUserHandler(cfg)))
	mux.Handle("POST /api/chirps", middlewareIsAuthenticated(cfg, createChirpHandler(cfg)))
	mux.Handle("GET /api/chirps", getChirpsHandler(cfg))
	mux.Handle("GET /api/chirps/{chirpID}", getChirpByIDHandler(cfg))
	mux.Handle("POST /api/login", loginHandler(cfg))
	mux.Handle("POST /api/refresh", refreshHandler(cfg))
	mux.Handle("POST /api/revoke", revokeHandler(cfg))

	log.Printf("Server started at %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
