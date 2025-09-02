package main

import (
	"sync/atomic"
	"fmt"
	"net/http"
	"github.com/paul39-33/chirpy/internal/database"
	"github.com/google/uuid"
	"time"
	"encoding/json"
	"log"
)

//struct to keep track of number of requests
type apiConfig struct {
	fileserverHits	atomic.Int32
	dbQueries		*database.Queries
	platform		string
}

//struct for user json data
type User struct {
	ID 			uuid.UUID `json:"id"`
	CreatedAt	time.Time `json:"created_at"`
	UpdatedAt	time.Time `json:"updated_at"`
	Email		string `json:"email"`
}

//increments fileserverHits every time its called
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		cfg.fileserverHits.Add(int32(1)) //middleware does its job
		next.ServeHTTP(w, r) //needs to use ServeHTTP to continue the http.FileServer
	})
}


//handler to show number of fileserverHits
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/html")
	metricsText := fmt.Sprintf(`
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
	w.Write([]byte(metricsText))
}


//handler to reset the fileserverHits count and delete all users
func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request){
	//check if the reset is being done by a dev
	if cfg.platform	!= "dev"{
		respondWithError(w, 403, "Command must be done by a dev")
		return
	}
	
	//call func to delete users
	if err := cfg.dbQueries.ResetUser(r.Context()); err != nil {
		log.Printf("Error decoding request: %v", err)
		respondWithError(w, 400, "Error resetting user")
		return
	}

	//cfg.fileserverHits.Store(0)
	w.Write([]byte("Reset successful"))
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request){
	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding request: %v", err)
		respondWithError(w, 400, "Error decoding request")
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error decoding request: %v", err)
		respondWithError(w, 400, "Error creating user")
		return
	}

	createdUser := User {
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}

	respondWithJSON(w, 201, createdUser)
}