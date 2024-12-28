package utils

import (
	"strings"

	_ "github.com/lib/pq"
)

var profaneWords = map[string]bool{
	"kerfuffle": true,
	"sharbert":  true,
	"fornax":    true,
}

func RemoveProfanity(body string) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		lowerCaseWord := strings.ToLower(word)
		if profaneWords[lowerCaseWord] {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}
