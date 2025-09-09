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
	"unicode/utf8"
	"errors"
	"database/sql"
	"github.com/paul39-33/chirpy/internal/auth"
)

//struct to keep track of number of requests
type apiConfig struct {
	fileserverHits	atomic.Int32
	dbQueries		*database.Queries
	platform		string
	secret			string
}

//struct for userlogin json data
type UserLogin struct {
	ID 				uuid.UUID `json:"id"`
	CreatedAt		time.Time `json:"created_at"`
	UpdatedAt		time.Time `json:"updated_at"`
	Email			string `json:"email"`
	IsChirpyRed		bool `json:"is_chirpy_red"`
	Token			string `json:"token"`
	RefreshToken	string `json:"refresh_token"`
}

type User struct {
	ID 				uuid.UUID `json:"id"`
	CreatedAt		time.Time `json:"created_at"`
	UpdatedAt		time.Time `json:"updated_at"`
	Email			string `json:"email"`
	IsChirpyRed		bool `json:"is_chirpy_red"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
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
		Password 	string `json:"password"`
		Email 		string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding request: %v", err)
		respondWithError(w, 400, "Error decoding request")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		respondWithError(w, 400, "Error hashing password")
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		HashedPassword: hashedPassword,
		Email: params.Email,
	})
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

func (cfg *apiConfig) handlerCreateChirps(w http.ResponseWriter, r *http.Request){
	//get user token
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting user token: %v", err)
		respondWithError(w, 400, "Error validating token")
		return
	}
	//validate user token
	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		log.Printf("Error validating user token: %v", err)
		respondWithError(w, 401, "Invalid token session")
		return
	}

	type parameters struct {
		Body	string		`json:"body"`
	}
	params := parameters{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding params.body: %v", err)
		respondWithError(w, 400, "Error decoding json")
		return
	}

	//count the length of the Body characters
	runeCount := utf8.RuneCountInString(params.Body)
	if runeCount == 0 || runeCount > 140 {
		log.Printf("Invalid body length!")
		respondWithError(w, http.StatusBadRequest, "Invalid chirp input!")
		return
	}

	//clean profanity texts
	params.Body = cleanProfanity(params.Body)

	chirp, err := cfg.dbQueries.CreateChirps(r.Context(), database.CreateChirpsParams{
		Body: params.Body,
		UserID: userID,
	})
	if err != nil {
		log.Printf("Error creating chirp: %v", err)
		respondWithError(w, 400, "Error creating chirp")
		return
	}
	createdChirp := Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	}

	respondWithJSON(w, 201, createdChirp)
}

//get all chirps
func(cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request){
	getChirps, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {
		log.Printf("Error getting chirps: %v", err)
		respondWithError(w, 400, "Error getting chirps")
		return
	}
	
	resp := make([]Chirp, len(getChirps))
	for i, c := range getChirps {
		resp[i] = Chirp{
			ID:	c.ID,
			CreatedAt:	c.CreatedAt,
			UpdatedAt:	c.UpdatedAt,
			Body:	c.Body,
			UserID:	c.UserID,
		}
	}
	respondWithJSON(w, 200, resp)
}

//get specific chirp by id
func(cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request){
	chirpID := r.PathValue("chirpID")
	
	//parse ID from string to uuid
	id, err := uuid.Parse(chirpID)
	if err != nil {
		log.Printf("Error parsing chirp ID from string to UUID: %v", err)
		respondWithError(w, 400, "Error parsing chirp ID")
		return
	}

	chirp, err := cfg.dbQueries.GetChirp(r.Context(), id)
	//if the error is because no matching chirp is found
	if errors.Is(err, sql.ErrNoRows){
		log.Printf("No matching chirp found: %v", err)
		respondWithError(w, 404, "Matching chirp not found")
		return
	}
	if err != nil {
		log.Printf("Error getting chirp by id: %v", err)
		respondWithError(w, 400, "Error getting chirp")
		return
	}

	resp := Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	}

	respondWithJSON(w, 200, resp)
}

func(cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request){
	type parameters struct{
		Password			string 		`json:"password"`
		Email				string 		`json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding request: %v", err)
		respondWithError(w, 400, "Error decoding request")
		return
	}

	user, err := cfg.dbQueries.UserLogin(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error getting password from database: %v", err)
		respondWithError(w, 400, "Error getting user")
		return
	}

	//compare pwFromDatabase with the password input
	if err = auth.CheckPasswordHash(params.Password, user.HashedPassword); err != nil {
		log.Printf("Incorrect email or password")
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	//access token expire duration
	accessTokenExp := 1 *time.Hour
	//create an access token after successful login
	token, err := auth.MakeJWT(user.ID, cfg.secret, accessTokenExp)
	if err != nil {
		log.Printf("Error creating access token: %v", err)
		respondWithError(w, 400, "Error creating access token")
		return
	}

	//create a refresh token with 60 days expiration time
	refreshTokenExp := (60 * 24) *time.Hour

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error creating refresh token: %v", err)
		respondWithError(w, 400, "Error creating refresh token")
		return
	}

	//store refresh token in DB
	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token: refreshToken,
		UserID: user.ID,
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(refreshTokenExp),
	})

	userInfo := UserLogin{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
		IsChirpyRed: user.IsChirpyRed,
		Token: token,
		RefreshToken: refreshToken,
	}

	respondWithJSON(w, 200, userInfo)
}

func(cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request){
	//get refresh token from token bearer
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting refresh token from header: %v", err)
		respondWithError(w, 400, "Error getting refresh token info")
		return
	}

	info, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), refreshToken)
	//if the error is because no result found
	if errors.Is(err, sql.ErrNoRows){
		log.Printf("No matching refresh token found: %v", err)
		respondWithError(w, 401, "Error invalid token")
		return
	}
	if err != nil{
		log.Printf("Error getting user from refresh token: %v", err)
		respondWithError(w, 500, "Error getting user info")
		return
	}

	//check if the refresh token has expired (revoked_at becomes NOT NULL = valid)
	if info.RevokedAt.Valid || time.Now().After(info.ExpiresAt){
		log.Printf("Refresh token already expired!")
		respondWithError(w, 401, "refresh token expired")
		return
	}

	//create new access token
	new_token, err := auth.MakeJWT(info.UserID, cfg.secret, time.Hour)
	if err != nil {
		log.Printf("Error creating new access token: %v", err)
		respondWithError(w, 400, "trouble creating new access token")
		return
	}

	type response struct {
		Token string `json:"token"`
	}

	resp := response{Token: new_token}

	respondWithJSON(w, 200, resp)
}

func(cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request){
	//get refresh token from token bearer
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting refresh token from header: %v", err)
		respondWithError(w, 400, "Error getting refresh token info")
		return
	}

	//call query to update the token in database to be revoked
	err = cfg.dbQueries.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		log.Printf("Error revoking refresh token: %v", err)
		respondWithJSON(w, 400, "Error revoking refresh token")
		return
	}

	respondWithJSON(w, 204, "refresh token has been revoked")
}

func(cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	//validate user
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting user token: %v", err)
		respondWithError(w, 401, "Error validating token")
		return
	}
	//validate user token
	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		log.Printf("Error validating user token: %v", err)
		respondWithError(w, 401, "Invalid token session")
		return
	}

	//decode the new email and password input
	type parameters struct{
		Password	string 	`json:"password"`
		Email		string 	`json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding request: %v", err)
		respondWithError(w, 400, "Error decoding request")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		respondWithError(w, 400, "Error hashing password")
		return
	}

	//update the user data in database
	err = cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		HashedPassword: hashedPassword,
		Email: params.Email,
		ID: userID,
	})
	if err != nil {
		log.Printf("Error updating user data : %v", err)
		respondWithError(w, 400, "Error updating user data")
		return
	}

	//get the new user data to print
	userInfo, err := cfg.dbQueries.UserLogin(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error getting user data: %v", err)
		respondWithError(w, 400, "Error retrieving user data")
		return
	}

	resp := User{
		ID: userInfo.ID,
		Email: userInfo.Email,
		IsChirpyRed: userInfo.IsChirpyRed,
	}

	respondWithJSON(w, 200, resp)
}

func(cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	//validate user
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting user token: %v", err)
		respondWithError(w, 401, "Error validating token")
		return
	}
	//validate user token
	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		log.Printf("Error validating user token: %v", err)
		respondWithError(w, 401, "Invalid token session")
		return
	}

	chirpID := r.PathValue("chirpID")
	
	//parse ID from string to uuid
	id, err := uuid.Parse(chirpID)
	if err != nil {
		log.Printf("Error parsing chirp ID from string to UUID: %v", err)
		respondWithError(w, 400, "Error parsing chirp ID")
		return
	}

	chirp, err := cfg.dbQueries.GetChirp(r.Context(), id)
	//if the error is because no matching chirp is found
	if errors.Is(err, sql.ErrNoRows){
		log.Printf("No matching chirp found: %v", err)
		respondWithError(w, 404, "Matching chirp not found")
		return
	}
	if err != nil {
		log.Printf("Error getting chirp by id: %v", err)
		respondWithError(w, 400, "Error getting chirp")
		return
	}

	//check if user ID is the same as the chirp's creator
	if chirp.UserID != userID {
		log.Printf("User has no access to chirp!")
		respondWithError(w, 403, "chirp access forbidden")
		return
	}

	err = cfg.dbQueries.DeleteChirp(r.Context(), id)
	if err != nil {
		log.Printf("Error deleting chirp: %v", err)
		respondWithError(w, 400, "problem deleting chirp")
		return
	}

	respondWithJSON(w, 204, "chirp removed")
}

func(cfg *apiConfig) handlerUpgradeUser(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		UserID	string `json:"user_id"`
	}
	
	type parameters struct {
		Event	string	`json:"event"`
		Data	Data	`json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding request: %v", err)
		respondWithError(w, 400, "Error decoding request")
		return
	}

	if params.Event != "user.upgraded" {
		respondWithJSON(w, 204, "")
		return
	}

	
	//parse user ID from string to uuid
	id, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		log.Printf("Error parsing user ID from string to UUID: %v", err)
		respondWithError(w, 400, "Error parsing user ID")
		return
	}

	err = cfg.dbQueries.UpgradeUser(r.Context(), id)
	//if the error is because no matching user ID is found
	if errors.Is(err, sql.ErrNoRows){
		log.Printf("No matching user ID found: %v", err)
		respondWithError(w, 404, "user id not found")
		return
	}
	if err != nil {
		log.Printf("error upgrading user: %v", err)
		respondWithError(w, 404, "error upgrading user")
		return
	}

	respondWithJSON(w, 204, "")
}