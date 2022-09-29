package api

import (
	"fmt"
	"humanitec.io/custom-reference-driver/internal/version"
	"net/http"

	"github.com/gorilla/mux"
)

// MapRoutes maps HTTP routes
func (api *apiServer) MapRoutes(router *mux.Router) error {

	// Public Routes
	router.Methods("PUT").Path("/s3/{GUResID}").HandlerFunc(api.upsertS3)
	router.Methods("DELETE").Path("/s3/{GUResID}").HandlerFunc(api.deleteS3)

	// Static & Service Routes
	router.Methods("GET").Path("/docs/spec.json").HandlerFunc(api.apiSpec)
	router.Methods("GET").Path("/alive").HandlerFunc(api.isAlive)
	router.Methods("GET").Path("/health").HandlerFunc(api.isReady)

	return nil
}

func (api *apiServer) apiSpec(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./openapi/spec.json")
}

func (api *apiServer) isAlive(w http.ResponseWriter, _ *http.Request) {
	writeAsText(w, http.StatusOK, fmt.Sprintf("%s %s (build: %s; sha: %s)", api.Name, version.Version, version.BuildTime, version.GitSHA))
}

func (api *apiServer) isReady(w http.ResponseWriter, _ *http.Request) {
	writeAsJSON(w, http.StatusOK, map[string]string{
		"app":        api.Name,
		"version":    version.Version,
		"build_time": version.BuildTime,
		"git_sha":    version.GitSHA,
		"status":     "OK",
	})
}
