package api

import (
	"github.com/gorilla/mux"
	"github.com/humanitec/golib/hlogger"

	"humanitec.io/go-service-template/internal/model"
)

// Server contains routes, handlers, configuration for the API server
type Server interface {
	MapRoutes(router *mux.Router) error
}

// NewServer initializes new Server instance
func NewServer(
	appName string,

	// TODO: Add your server attributes here

	db model.Databaser,
	logger *hlogger.HLogger,
) Server {
	return &apiServer{
		Name: appName,

		databaser: db,
		logger:    logger,
	}
}

type apiServer struct {
	Name string

	// TODO: Add your server attributes here

	databaser model.Databaser
	logger    *hlogger.HLogger
}
