package main

import (
	"net/http"
	"log"
	_ "github.com/lib/pq"
	"os"
	"github.com/paul39-33/chirpy/internal/database"
	"database/sql"
	"github.com/joho/godotenv"
)




func main(){
	//load .env file to environment variables
	godotenv.Load()
	//get db url from .env file
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error connecting to sql: %v", err)
	}
	dbQueries := database.New(db)

	platform := os.Getenv("PLATFORM")
	mux := http.NewServeMux()
	apiCfg := apiConfig{
		dbQueries: dbQueries,
		platform: platform,
	}

	//create a server variable
	srv := http.Server{
		Handler:	mux,
		Addr:		":8080",
	}	

	file_srv := http.StripPrefix("/app/",http.FileServer(http.Dir(".")))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(file_srv))

	mux.HandleFunc("GET /api/healthz", handlerHealthz)
	
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)

	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirps)

	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)

	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirp)

	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)

	

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

}


