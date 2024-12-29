package main

import (
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

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func createUserHandler(cfg *config.ApiConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			log.Printf("Error decoding parameters: %v", err)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
			return
		}
		hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		user, err := cfg.Queries.CreateUser(r.Context(), database.CreateUserParams{
			Email:          params.Email,
			HashedPassword: hashedPassword,
		})
		if err != nil {
			log.Printf("Error creating user: %v", err)
			utils.RespondWithError(w, http.StatusBadRequest, "Email already in use")
			return
		}

		utils.RespondWithJSON(w, http.StatusCreated, User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		})
	})
}

func updateUserHandler(cfg *config.ApiConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			log.Printf("Error decoding parameters: %v", err)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		userID := r.Context().Value(userIDKey).(uuid.UUID)
		hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		user, err := cfg.Queries.UpdateUser(
			r.Context(),
			database.UpdateUserParams{
				ID:             userID,
				Email:          params.Email,
				HashedPassword: hashedPassword,
			},
		)
		if err != nil {
			log.Printf("Error updating user: %v", err)
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		})
	})
}
