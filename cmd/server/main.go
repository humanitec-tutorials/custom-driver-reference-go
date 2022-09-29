package main

import (
	"flag"
	"fmt"
	"github.com/spf13/cast"
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

	// Get host
	host := flag.String("h", "''", "The ip to listen for incoming requests on (`''` = accept all).")
	// Get port
	port := flag.Int("p", 8080, "The port to listen for the incoming requests on (default is 8080).")
	// Get port
	log_level := flag.String("l", "info", "The level of logging expected (`'info'`,`'warn'`,`'error'`,`'debug'`).")
	// Get port
	use_fake_aws := flag.Bool("m", false, "Use the mock AWS API (for unitests). ")

	flag.Usage = func() {
		flagSet := flag.CommandLine
		fmt.Printf("Custom Usage of %s:\n", "./server")
		order := []string{"h", "p", "l", "m"}
		for _, name := range order {
			flag := flagSet.Lookup(name)
			fmt.Printf("-%s\n", flag.Name)
			fmt.Printf("  %s\n", flag.Usage)
		}
	}

	flag.Parse()

	conf.Host = *host
	conf.Port = cast.ToInt(port)
	conf.LogLevel = *log_level
	conf.FakeAWSClient = cast.ToBool(use_fake_aws)

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

	log.Println("Starting server", "host", *host, "port", cast.ToString(port))

	if err := http.ListenAndServe(*host+":"+cast.ToString(port), router); err != nil {
		log.Fatal("Failed starting server", "err", err)
	}
}
