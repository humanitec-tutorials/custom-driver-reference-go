package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/humanitec/golib/hlogger"
	"github.com/stretchr/testify/assert"

	"humanitec.io/go-service-template/internal/model"
	"humanitec.io/go-service-template/internal/testutil"

	db_mock "humanitec.io/go-service-template/internal/model/mocks"
)

func stringPtr(str string) *string {
	return &str
}

func TestGetEntity(t *testing.T) {
	const (
		testEntityID = "test-entity"
		testUserID   = "test-user"
	)

	var (
		timestamp   = time.Now().UTC()
		savedEntity = model.Entity{
			ID:        testEntityID,
			Name:      "Test Entity",
			Status:    "Active",
			CreatedBy: testUserID,
			CreatedAt: &timestamp,
		}
		expectedEntity = Entity{
			ID:        testEntityID,
			Name:      "Test Entity",
			Status:    "Active",
			CreatedBy: testUserID,
			CreatedAt: &timestamp,
		}
	)

	var tests = []struct {
		Name       string
		EntityID   *string
		FromHeader *string

		DbData  *model.Entity
		DbError error

		ExpectedStatus int
		ExpectedOutput *Entity
		ExpectedError  error
	}{
		// Success path
		//
		{
			Name:           "Should return the entity details",
			DbData:         &savedEntity,
			ExpectedStatus: http.StatusOK,
			ExpectedOutput: &expectedEntity,
		},

		// Errors handling
		//
		{
			Name:           "Should return HTTP 400 for invalid entity ID",
			EntityID:       stringPtr("__"),
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError:  errors.New("Invalid entity ID"),
		},
		{
			Name:           "Should return HTTP 401 for missing or empty HTTP From header",
			FromHeader:     stringPtr(""),
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "Should return HTTP 404 for non existent entity",
			DbError:        model.ErrNotFound,
			ExpectedStatus: http.StatusNotFound,
		},
		{
			Name:           "Should return HTTP 500 error for DB connectivity errors",
			DbError:        errors.New("test error"),
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedError:  errors.New("Technical issue"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Prepare request data (apply default tests conditions)
			//

			var entityID string = testEntityID
			if tt.EntityID != nil {
				entityID = *tt.EntityID
			}

			headers := map[string]string{
				"Accept": "application/json",
				"From":   testUserID,
			}
			if tt.FromHeader != nil {
				headers["From"] = *tt.FromHeader
			}

			// Setup mocks
			//

			db := db_mock.NewMockDatabaser(ctrl)
			db.
				EXPECT().
				SelectEntity(testutil.WithTestContext(), entityID).
				Return(tt.DbData, tt.DbError).
				MaxTimes(1)

			logger, _ := hlogger.NewTestLogger()
			api := &apiServer{
				databaser: db,
				logger:    logger,
			}

			// Initialize test server
			//

			router := mux.NewRouter()
			api.MapRoutes(router)

			// Execute test request
			//

			url := fmt.Sprintf("/entities/%s", entityID)
			resp, err := executeTestGet(testutil.TestContext(), router, url, headers)
			assert.NoError(t, err, "Failed to send test request")

			// Validate test results
			//

			assert.Equal(t, tt.ExpectedStatus, resp.Code,
				"Should return HTTP %v. Actual: HTTP %v", tt.ExpectedStatus, resp.Code)

			if tt.ExpectedError != nil {
				// On Error
				//
				assert.Contains(t, resp.Body.String(), tt.ExpectedError.Error())
			} else if tt.ExpectedOutput != nil {
				// On success
				//
				var resData Entity
				var err = json.Unmarshal(resp.Body.Bytes(), &resData)
				assert.NoError(t, err, "Response data should be a valid JSON. Actual: %s", resp.Body)

				assert.Equal(t, &resData, tt.ExpectedOutput)
			}
		})
	}
}
