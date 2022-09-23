package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"humanitec.io/go-service-template/internal/version"
)

// MapRoutes maps HTTP routes
func (api *apiServer) MapRoutes(router *mux.Router) error {

	// Public Routes
	router.Methods("GET").Path("/entities/{id}").HandlerFunc(api.getEntity)

	// Internal Routes
	router.Methods("GET").Path("/internal/entities/{id}").HandlerFunc(api.getEntity)

	// Static & Service Routes
	router.Methods("GET").Path("/docs/spec.json").HandlerFunc(api.apiSpec)
	router.Methods("GET").Path("/alive").HandlerFunc(api.isAlive)
	router.Methods("GET").Path("/health").HandlerFunc(api.isReady)

	return nil
}

func (api *apiServer) apiSpec(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./openapi/spec.json")
}

func (api *apiServer) isAlive(w http.ResponseWriter, r *http.Request) {
	api.writeAsText(r.Context(), w, http.StatusOK, fmt.Sprintf("%s %s (build: %s; sha: %s)", api.Name, version.Version, version.BuildTime, version.GitSHA))
}

func (api *apiServer) isReady(w http.ResponseWriter, r *http.Request) {
	api.writeAsJSON(r.Context(), w, http.StatusOK, map[string]string{
		"app":        api.Name,
		"version":    version.Version,
		"build_time": version.BuildTime,
		"git_sha":    version.GitSHA,
		"status":     "OK",
	})
}
