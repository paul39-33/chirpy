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

	mux.Handle("/", file_srv)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

}