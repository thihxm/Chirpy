package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	t.Run("hashes password", func(t *testing.T) {
		password := "password"
		hashedPassword, err := HashPassword(password)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if hashedPassword == password {
			t.Errorf("expected hashed password, got original password")
		}
	})
}

func TestComparePassword(t *testing.T) {
	t.Run("compares password", func(t *testing.T) {
		password := "password"
		hashedPassword, err := HashPassword(password)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		err = ComparePassword(hashedPassword, password)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("errors on invalid password", func(t *testing.T) {
		password := "password"
		hashedPassword, err := HashPassword(password)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		err = ComparePassword(hashedPassword, "invalid-password")
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestMakeJWT(t *testing.T) {
	t.Run("makes JWT", func(t *testing.T) {
		userID := uuid.New()
		tokenSecret := "token-secret"
		expiresIn := 1 * time.Minute
		token, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if token == "" {
			t.Errorf("expected token, got empty string")
		}
	})
}

func TestValidateJWT(t *testing.T) {
	t.Run("validates JWT", func(t *testing.T) {
		userID := uuid.New()
		tokenSecret := "token-secret"
		expiresIn := 1 * time.Minute
		token, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		validUserID, err := ValidateJWT(token, tokenSecret)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if validUserID != userID {
			t.Errorf("expected: %s, got: %s", userID, validUserID)
		}
	})
}

func TestValidateJWT_ErrorsOnInvalidToken(t *testing.T) {
	t.Run("errors on invalid token", func(t *testing.T) {
		token := "invalid"
		tokenSecret := "token-secret"
		_, err := ValidateJWT(token, tokenSecret)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestValidateJWT_ErrorsOnInvalidTokenSecret(t *testing.T) {
	t.Run("errors on invalid token secret", func(t *testing.T) {
		userID := uuid.New()
		tokenSecret := "token-secret"
		expiresIn := 1 * time.Minute
		token, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		_, err = ValidateJWT(token, "invalid-token-secret")
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestValidateJWT_ErrorsOnExpiredToken(t *testing.T) {
	t.Run("errors on expired token", func(t *testing.T) {
		userID := uuid.New()
		tokenSecret := "token-secret"
		expiresIn := 20 * time.Millisecond
		token, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		time.Sleep(50 * time.Millisecond)

		_, err = ValidateJWT(token, tokenSecret)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestGetBearerToken_ValidInput(t *testing.T) {
	var tests = []struct {
		name     string
		headers  http.Header
		expected string
	}{
		{
			"valid header",
			http.Header{"Authorization": []string{"Bearer token"}},
			"token",
		},
		{
			"another valid header",
			http.Header{"Authorization": []string{"Bearer another-token"}},
			"another-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GetBearerToken(tt.headers)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if token != tt.expected {
				t.Errorf("expected: %s, got: %s", tt.expected, token)
			}
		})
	}
}
func TestGetBearerToken_ErrorsOnInvalidInput(t *testing.T) {
	var tests = []struct {
		name     string
		headers  http.Header
		expected string
	}{
		{
			"no header",
			http.Header{},
			"",
		},
		{
			"empty header",
			http.Header{"Authorization": []string{""}},
			"",
		},
		{
			"invalid header format (extra spaces)",
			http.Header{"Authorization": []string{"Bearer   token"}},
			"",
		},
		{
			"invalid header format (missing Bearer)",
			http.Header{"Authorization": []string{"token"}},
			"",
		},
		{
			"invalid header format (wrong type)",
			http.Header{"Authorization": []string{"Basic token"}},
			"",
		},
		{
			"invalid header format (extra data)",
			http.Header{"Authorization": []string{"Basic token extra"}},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetBearerToken(tt.headers)
			if err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
