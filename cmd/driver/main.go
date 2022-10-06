package main

import (
	"github.com/gorilla/mux"
	"humanitec.io/custom-reference-driver/internal"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	log.Println("Starting server http://0.0.0.0:8080")

	router := mux.NewRouter()
	internal.MapRoutes(router)

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal("Failed starting server", "err", err)
	}
}
