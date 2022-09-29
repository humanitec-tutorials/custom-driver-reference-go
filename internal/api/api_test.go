package api

import (
	"humanitec.io/custom-reference-driver/internal/testutils"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestRouter(t *testing.T) {
	s := &apiServer{}
	router := mux.NewRouter()
	assert.NoError(t, s.MapRoutes(router))
	assert.NoError(t, router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		routeErr := route.GetError()
		if routeErr != nil {
			return routeErr
		}
		return nil
	}))
}

func TestStaticRoutes(t *testing.T) {
	// As tests run from the current dir we need this trick to change dir as if they run from the module's root
	// otherwise it can't read swagger files
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../..")
	os.Chdir(dir)

	s := &apiServer{}
	router := mux.NewRouter()
	assert.NoError(t, s.MapRoutes(router))
	rr := testutils.ExecuteTestRequest(testutils.TestContext(), t, router, "GET", "/health", nil, nil)
	assert.Equal(t, rr.Code, http.StatusOK)
}

func TestOpenAPISpecRoute(t *testing.T) {
	// As tests run from the current dir we need this trick to change dir as if they run from the module's root
	// otherwise it can't read swagger files
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../..")
	os.Chdir(dir)
	dat, err := ioutil.ReadFile("openapi/spec.json")
	assert.NoError(t, err)
	s := &apiServer{}
	router := mux.NewRouter()
	assert.NoError(t, s.MapRoutes(router))
	rr := testutils.ExecuteTestRequest(testutils.TestContext(), t, router, "GET", "/docs/spec.json", nil, nil)
	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Equal(t, string(dat), rr.Body.String())
}
