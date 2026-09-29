package main

import (
	"log"
	"net/http"
)

func main() {

	mux := http.ServeMux{}

	server := http.Server{
		Addr:    ":8080",
		Handler: &mux,
	}

	log.Fatal(server.ListenAndServe())

}
