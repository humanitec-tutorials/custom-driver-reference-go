package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/humanitec/golib/hlogger"
	"github.com/humanitec/golib/httplogger"
	"humanitec.io/go-service-template/internal/api"
	"humanitec.io/go-service-template/internal/config"
	"humanitec.io/go-service-template/internal/model"
	"humanitec.io/go-service-template/internal/version"

	muxtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"
	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	AppName = "go-service-template"
)

// serviceString can be used to log service details before the logger is initialized
func serviceString() string {
	return fmt.Sprintf("%s %s (build: %s; sha: %s)", AppName, version.Version, version.BuildTime, version.GitSHA)
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	// Logging
	logger, err := hlogger.NewHLogger(cfg.LogLevel, false, "json")
	if err != nil {
		log.Fatalf("Error building logger: %v (%s)", err, serviceString())
	}
	defer hlogger.OnExit(logger.Logger)
	sugar := logger.Logger.Sugar()

	sugar.Infow("Starting", "app", AppName, "version", version.Version, "build", version.BuildTime, "sha", version.GitSHA)

	if cfg.DataDogEnabled {
		ddtrace.Start(ddtrace.WithServiceVersion(version.Version))
		defer ddtrace.Stop()
	}

	dbConnStr := fmt.Sprintf(
		"dbname=%s user=%s password=%s host=%s port=%s connect_timeout=1 sslmode=disable",
		cfg.DatabaseName, cfg.DatabaseUser, cfg.DatabasePassword, cfg.DatabaseHost, cfg.DatabasePort)
	db, err := model.NewDatabaser(context.Background(), logger, dbConnStr, 6, true)
	if err != nil {
		sugar.Fatalw("Initializing database", "err", err)
	}

	// TODO: Place additional initialization here

	s := api.NewServer(
		AppName,
		db,
		logger,
	)

	router := muxtrace.NewRouter()
	if err := s.MapRoutes(router.Router); err != nil {
		sugar.Fatalw("Unable to initialize API server routes", "err", err)
	}
	router.Use(httplogger.LoggingMiddleware(&httplogger.Config{
		Logger:        logger.Logger,
		SilencedPaths: []string{"/alive", "/health"},
	}))

	// Get port
	port := flag.String("p", "8080", "Port (default is 8080)")
	flag.Parse()

	sugar.Infow("Starting server", "port", *port)

	if err := http.ListenAndServe(":"+*port, router); err != nil {
		sugar.Fatalw("Failed starting server", "err", err)
	}
}
