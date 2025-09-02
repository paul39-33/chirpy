package main

import (
	"slices"
	"strings"
)

var profanityTexts = []string{"kerfuffle", "sharbert", "fornax"}

func cleanProfanity(text string) string{
	textFields := strings.Fields(text)
	for i,text := range textFields{
		if slices.Contains(profanityTexts, strings.ToLower(text)){
			textFields[i] = "****"
		}
	}
	joinText := strings.Join(textFields, " ")
	return joinText
}