package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"humanitec.io/go-service-template/internal/model"
)

// oas:schema
// An Entity.
type Entity struct {

	// The unique ID for the Entity.
	// oas:pattern ^[a-z0-9][a-z0-9-]+[a-z0-9]$
	ID string `json:"id"`

	// The human friendly name for the Entity.
	Name string `json:"name"`

	// The Entity status.
	Status string `json:"status"`

	// The timestamp when the Entity was created.
	// oas:onlyin response
	CreatedAt *time.Time `json:"created_at"`

	// The User ID who has created the Entity.
	// oas:onlyin response
	CreatedBy string `json:"created_by"`
}

// oas:operation get /entities/{id}; Returns the Entity details.
//
// oas:param id; The Entity ID.
//
// oas:response 200 Entity; Rhe Entity details.
// oas:response 400; Invalid request parameters or payload. E.g. invalid `id` format.
// oas:response 401; Required HTTP Authorization header is missing or malformed.
// oas:response 500; Internal application error.
//
// oas:tags public Entity
func (api *apiServer) getEntity(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()

	api.logger.Logger.Sugar().Infof("getEntity: From: '%v'", r.Header["from"])
	_, err := getUser(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		api.writeError(ctx, w, http.StatusUnauthorized, "Unauthorized",
			fmt.Errorf("%s %s: %w\n", r.Method, r.URL, err))
		return
	}

	params := mux.Vars(r)
	id := params["id"]

	if !isValidAsID(id) {
		api.writeError(ctx, w, http.StatusBadRequest, "Invalid entity ID", nil)
		return
	}

	entity, err := api.databaser.SelectEntity(ctx, id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		api.writeError(ctx, w, http.StatusInternalServerError, "Technical issue", err)
		return
	}

	res := Entity{
		ID:        entity.ID,
		Name:      entity.Name,
		Status:    entity.Status,
		CreatedAt: entity.CreatedAt,
		CreatedBy: entity.CreatedBy,
	}
	api.writeAsJSON(ctx, w, http.StatusOK, res)
}
