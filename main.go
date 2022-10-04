package main

import (
	"github.com/gorilla/mux"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	AppName = "custom-reference-driver"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	router := mux.NewRouter()

	// Public Routes
	router.Methods("PUT").Path("/s3/{GUResID}").HandlerFunc(upsertS3)
	router.Methods("DELETE").Path("/s3/{GUResID}").HandlerFunc(deleteS3)

	// Static & Service Routes
	router.Methods("GET").Path("/docs/spec.json").HandlerFunc(apiSpec)
	router.Methods("GET").Path("/alive").HandlerFunc(isAlive)
	router.Methods("GET").Path("/health").HandlerFunc(isReady)

	log.Println("Starting server http://0.0.0.0:8080")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal("Failed starting server", "err", err)
	}
}
