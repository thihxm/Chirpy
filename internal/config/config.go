package config

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/thihxm/Chirpy/internal/database"
	"github.com/thihxm/Chirpy/internal/utils"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	Queries        *database.Queries
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) Metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(
		`<html>
        <body>
            <h1>Welcome, Chirpy Admin</h1>
            <p>Chirpy has been visited %d times!</p>
        </body>
        </html>`,
		cfg.FileserverHits.Load())),
	)
}

func (cfg *ApiConfig) Reset(w http.ResponseWriter, r *http.Request) {
	platform := os.Getenv("PLATFORM")
	if platform != "dev" {
		utils.RespondWithError(w, http.StatusForbidden, "Reset only allowed in dev environment")
		return
	}

	err := cfg.Queries.ResetUsers(r.Context())
	if err != nil {
		log.Printf("Error resetting users: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	cfg.FileserverHits.Store(0)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
