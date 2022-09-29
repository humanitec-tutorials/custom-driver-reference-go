package api

import (
	"github.com/gorilla/mux"
	"humanitec.io/custom-reference-driver/internal/aws"
)

// Server contains routes, handlers, configuration for the API server
type Server interface {
	MapRoutes(router *mux.Router) error
}

// NewServer initializes new Server instance
func NewServer(
	appName string,
	newAwsClient func(string, string, string) (aws.Client, error),

) Server {
	return &apiServer{
		Name:         appName,
		newAwsClient: newAwsClient,
	}
}

type apiServer struct {
	Name         string
	newAwsClient func(string, string, string) (aws.Client, error)
}
