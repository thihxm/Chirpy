package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/thihxm/Chirpy/internal/config"
	"github.com/thihxm/Chirpy/internal/database"
	"github.com/thihxm/Chirpy/internal/utils"
)

const (
	MaxChirpLength = 140
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func createChirpHandler(cfg *config.ApiConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		userID := r.Context().Value(userIDKey).(uuid.UUID)

		chirp, err := cfg.Queries.CreateChirp(r.Context(), database.CreateChirpParams{
			Body:   utils.RemoveProfanity(params.Body),
			UserID: userID,
		})
		if err != nil {
			log.Printf("Error creating user: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		utils.RespondWithJSON(w, http.StatusCreated, Chirp(chirp))
	})
}

func getChirpsHandler(cfg *config.ApiConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryAuthorID := r.URL.Query().Get("author_id")
		querySort := r.URL.Query().Get("sort")

		if querySort != "" && querySort != "asc" && querySort != "desc" {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid sort parameter")
		}
		if querySort == "" {
			querySort = "asc"
		}

		var authorID uuid.NullUUID
		if queryAuthorID != "" {
			parsedUUID, err := uuid.Parse(queryAuthorID)
			if err != nil {
				utils.RespondWithError(w, http.StatusBadRequest, "Invalid author ID")
				return
			}
			authorID = uuid.NullUUID{UUID: parsedUUID, Valid: true}
		}
		rawChirps, err := cfg.Queries.GetChirps(r.Context(), database.GetChirpsParams{
			AuthorID: authorID,
			Sort:     querySort,
		})

		if err != nil {
			log.Printf("Error getting chirps: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		chirps := make([]Chirp, len(rawChirps))
		for i, chirp := range rawChirps {
			chirps[i] = Chirp(chirp)
		}

		utils.RespondWithJSON(w, http.StatusOK, chirps)
	})
}

func getChirpByIDHandler(cfg *config.ApiConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawID := r.PathValue("chirpID")

		id, err := uuid.Parse(rawID)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
			return
		}

		chirp, err := cfg.Queries.GetChirpByID(r.Context(), id)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Chirp not found")
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, Chirp(chirp))
	})
}

func deleteChirpByIDHandler(cfg *config.ApiConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawID := r.PathValue("chirpID")

		id, err := uuid.Parse(rawID)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
			return
		}

		userID := r.Context().Value(userIDKey).(uuid.UUID)

		chirp, err := cfg.Queries.GetChirpByID(r.Context(), id)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Chirp not found")
			return
		}

		if chirp.UserID != userID {
			utils.RespondWithError(w, http.StatusForbidden, "This chirp does not belong to you")
			return
		}

		err = cfg.Queries.DeleteChirpByID(r.Context(), database.DeleteChirpByIDParams{
			ID:     id,
			UserID: userID,
		})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		utils.RespondWithJSON(w, http.StatusNoContent, nil)
	})
}
