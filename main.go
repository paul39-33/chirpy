package main

import (
	"net/http"
	"log"
	"fmt"
	"sync/atomic"
	"unicode/utf8"
	"encoding/json"
	"strings"
	"slices"
)

//struct to keep track of number of requests
type apiConfig struct {
	fileserverHits atomic.Int32
}

//json error struct
type returnErr struct {
	Error string `json:"error"`
}

//increments fileserverHits every time its called
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		cfg.fileserverHits.Add(int32(1)) //middleware does its job
		next.ServeHTTP(w, r) //needs to use ServeHTTP to continue the http.FileServer
	})
}

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

func main(){
	mux := http.NewServeMux()
	apiCfg := apiConfig{}

	//create a server variable
	srv := http.Server{
		Handler:	mux,
		Addr:		":8080",
	}	

	file_srv := http.StripPrefix("/app/",http.FileServer(http.Dir(".")))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(file_srv))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("POST /api/validate_chirp", func (w http.ResponseWriter, r *http.Request){
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
	})
	
	
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

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

//handler to reset the fileserverHits count
func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request){
	cfg.fileserverHits.Store(0)
	w.Write([]byte("Reset successful"))
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

