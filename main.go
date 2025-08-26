package main

import (
	"net/http"
	"log"
)

func main(){
	mux := http.NewServeMux()

	//create a server variable
	srv := http.Server{
		Handler:	mux,
		Addr:		":8080",
	}

	file_srv := http.FileServer(http.Dir("."))

	mux.Handle("/app/", http.StripPrefix("/app/", file_srv))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

}