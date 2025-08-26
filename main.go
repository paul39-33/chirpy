package main

import (
	"net/http"
	"log"
	"fmt"
	"sync/atomic"
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

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	mux.HandleFunc("/metrics", apiCfg.handlerMetrics)

	mux.HandleFunc("/reset", apiCfg.handlerReset)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

}

//handler to show number of fileserverHits
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request){
	metricsText := fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())
	w.Write([]byte(metricsText))
}

//handler to reset the fileserverHits count
func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request){
	cfg.fileserverHits.Store(0)
	w.Write([]byte("Reset successful"))
}