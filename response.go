package main

import (
	"encoding/json"
	"net/http"
)

//json error struct
type returnErr struct {
	Error string `json:"error"`
}

func respondWithError(w http.ResponseWriter, code int, msg string){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(returnErr{Error: msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}