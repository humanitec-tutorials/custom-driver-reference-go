package main

import (
	"flag"
	"fmt"
	"humanitec.io/custom-reference-driver/internal/api"
	"humanitec.io/custom-reference-driver/internal/aws"
	"humanitec.io/custom-reference-driver/internal/config"
	"humanitec.io/custom-reference-driver/internal/version"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const (
	AppName = "custom-reference-driver"
)

// serviceString can be used to log service details before the logger is initialized
func serviceString() string {
	return fmt.Sprintf("%s %s (build: %s; sha: %s)", AppName, version.Version, version.BuildTime, version.GitSHA)
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	conf, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	newAwsClient := aws.New
	if conf.FakeAWSClient {
		newAwsClient = aws.FakeNew
	}

	s := api.NewServer(
		AppName,
		newAwsClient,
	)

	router := mux.NewRouter()
	if err := s.MapRoutes(router); err != nil {
		log.Fatal("Unable to initialize API server routes", "err", err)
	}

	// Get port
	port := flag.String("p", "8080", "Port (default is 8080)")
	flag.Parse()

	log.Println("Starting server", "port", *port)

	if err := http.ListenAndServe(":"+*port, router); err != nil {
		log.Fatal("Failed starting server", "err", err)
	}
}
