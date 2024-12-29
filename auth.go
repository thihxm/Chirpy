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
	"github.com/thihxm/Chirpy/internal/database"
	"github.com/thihxm/Chirpy/internal/utils"
)

type contextKey string

const userIDKey contextKey = "userID"

const (
	DEFAULT_TOKEN_EXPIRATION_TIME = 1 * time.Hour
	MAX_TOKEN_EXPIRATION_TIME     = 1 * time.Hour
	REFRESH_TOKEN_EXPIRATION_TIME = 60 * 24 * time.Hour
)

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
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
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

		refreshToken, err := auth.MakeRefreshToken()
		if err != nil {
			log.Printf("Error generating refresh token: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		_, err = cfg.Queries.CreateRefreshToken(
			r.Context(),
			database.CreateRefreshTokenParams{
				Token:     refreshToken,
				UserID:    user.ID,
				ExpiresAt: time.Now().Add(REFRESH_TOKEN_EXPIRATION_TIME),
			},
		)
		if err != nil {
			log.Printf("Error creating refresh token: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, AuthenticatedUser{
			ID:           user.ID,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
			Email:        user.Email,
			Token:        token,
			RefreshToken: refreshToken,
		})
	})
}

func refreshHandler(cfg *config.ApiConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerRefreshToken, err := auth.GetBearerToken(r.Header)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		refreshToken, err := cfg.Queries.GetRefreshTokenByToken(r.Context(), headerRefreshToken)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		token, err := auth.MakeJWT(refreshToken.UserID, cfg.AuthSecret, DEFAULT_TOKEN_EXPIRATION_TIME)
		if err != nil {
			log.Printf("Error generating token: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		type resToken struct {
			Token string `json:"token"`
		}

		utils.RespondWithJSON(w, http.StatusOK, resToken{Token: token})
	})
}
