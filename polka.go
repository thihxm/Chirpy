package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/thihxm/Chirpy/internal/auth"
	"github.com/thihxm/Chirpy/internal/config"
	"github.com/thihxm/Chirpy/internal/database"
	"github.com/thihxm/Chirpy/internal/utils"
)

type webhookEventTypes string

const (
	userUpgraded webhookEventTypes = "user.upgraded"
)

func polkaWebhookHandler(cfg *config.ApiConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiKey, err := auth.GetAPIKey(r.Header); err != nil || apiKey != cfg.PolkaKey {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		type parameters struct {
			Event string `json:"event"`
			Data  struct {
				UserID uuid.UUID `json:"user_id"`
			} `json:"data"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			log.Printf("Error decoding parameters: %v", err)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		if params.Event != string(userUpgraded) {
			utils.RespondWithJSON(w, http.StatusNoContent, nil)
			return
		}

		_, err = cfg.Queries.UpgradeToChirpRed(
			r.Context(),
			database.UpgradeToChirpRedParams{
				ID:          params.Data.UserID,
				IsChirpyRed: true,
			},
		)
		if err != nil {
			log.Printf("Error upgrading user: %v", err)
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}

		utils.RespondWithJSON(w, http.StatusNoContent, nil)
	})
}
