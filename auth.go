package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/thihxm/Chirpy/internal/auth"
	"github.com/thihxm/Chirpy/internal/config"
	"github.com/thihxm/Chirpy/internal/utils"
)

type contextKey string

const userIDKey contextKey = "userID"

func middlewareIsAuthenticated(cfg *config.ApiConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		userID, err := auth.ValidateJWT(token, cfg.AuthSecret)
		if err != nil {
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		newReq := r.WithContext(ctx)
		next.ServeHTTP(w, newReq)
	})
}

type AuthenticatedUser struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func loginHandler(cfg *config.ApiConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Email            string `json:"email"`
			Password         string `json:"password"`
			ExpiresInSeconds *int   `json:"expires_in_seconds"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			log.Printf("Error decoding parameters: %v", err)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		expiresIn := DEFAULT_TOKEN_EXPIRATION_TIME

		if params.ExpiresInSeconds != nil {
			expiresIn = time.Duration(*params.ExpiresInSeconds) * time.Second

			if expiresIn < 0 || expiresIn > MAX_TOKEN_EXPIRATION_TIME {
				expiresIn = DEFAULT_TOKEN_EXPIRATION_TIME
			}
		}

		user, err := cfg.Queries.GetUserByEmail(r.Context(), params.Email)
		if err != nil {
			log.Printf("Error getting user: %v", err)
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}

		err = auth.ComparePassword(user.HashedPassword, params.Password)
		if err != nil {
			log.Printf("Error comparing password: %v", err)
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		token, err := auth.MakeJWT(user.ID, cfg.AuthSecret, expiresIn)
		if err != nil {
			log.Printf("Error generating token: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, AuthenticatedUser{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
			Token:     token,
		})
	})
}
