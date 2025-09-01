package main

import (
	"net/http"
	"log"
	"fmt"
	"sync/atomic"
	"unicode/utf8"
	"encoding/json"
	"strings"
)

//struct to keep track of number of requests
type apiConfig struct {
	fileserverHits atomic.Int32
}

//increments fileserverHits every time its called
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		cfg.fileserverHits.Add(int32(1)) //middleware does its job
		next.ServeHTTP(w, r) //needs to use ServeHTTP to continue the http.FileServer
	})
}

func cleanProfanity(text string) string{
	textToLower := strings.ToLower(text)
	
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

		type returnErr struct {
			Error string `json:"error"`
		}

		errBody := returnErr{
			Error: "Something went wrong",
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			dat, err := json.Marshal(errBody)
			if err != nil {
				log.Printf("Error marshaling json: %v", err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			w.Write(dat)
			return
		}

		//encoding response
		type returnVals struct {
			Valid bool `json:"valid"`
		}

		respBody := returnVals{
			Valid: true,
		}

		runeCount := utf8.RuneCountInString(params.Body)	
		if runeCount == 0 || runeCount > 140 {
			errRuneCount := returnErr{
				Error: "Chirp is too long",
			}

			dat, err := json.Marshal(errRuneCount)
			if err != nil {
				log.Printf("Error marshaling json: %v", err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			w.Write(dat)
			return
		}

		lowerInput := 

		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshaling json: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Something went wrong"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(dat)
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

