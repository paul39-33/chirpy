package main

import (
	"net/http"
	"encoding/json"
	"unicode/utf8"
)



func handlerHealthz(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request){
	//decoding request
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Error decoding request")
		return
	}

	//encoding response

	runeCount := utf8.RuneCountInString(params.Body)	
	if runeCount == 0 || runeCount > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	//clean profanity texts
	cleanedText := cleanProfanity(params.Body)
	respondWithJSON(w, http.StatusOK, map[string]string{
		"cleaned_body": cleanedText,
	})
}

