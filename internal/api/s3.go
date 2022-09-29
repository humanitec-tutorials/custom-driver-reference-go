package api

import (
	"fmt"
	"humanitec.io/custom-reference-driver/internal/cookies"
	"humanitec.io/custom-reference-driver/internal/errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// oas:operation PUT /s3/{GUResID}; Creates or Updates the resource.
//
// oas:param GUResID; The resource Id.
//
// oas:requestBody DriverInputs;
//
// oas:response 200 DriverOutputs; A resource data.
// oas:response 400 herrors.HumanitecError; Resource definition is not valid.
// oas:response 422 herrors.HumanitecError; The request body was not parsable.
func (api *apiServer) upsertS3(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	GUResID := params["GUResID"]
	ctx := r.Context()

	if !isValidAsID(GUResID) {
		writeError(w, http.StatusBadRequest, "GUResID is not a valid Humanitec ID", nil)
		return
	}

	// Parse and validate request payload
	//
	var inputs DriverInputs
	if err := readAsJSON(r.Body, &inputs); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "Unable to process the request", err)
		return
	} else if inputs.Type != "s3" {
		writeError(w, http.StatusBadRequest, "Invalid resource type", nil)
		return
	} else if inputs.Driver == nil {
		inputs.Driver = &ValuesSecrets{}
	}

	// Validate AWS access credentials
	//
	region := inputs.Driver.Values["region"].(string)
	accessKeyId, secretAccessKey, err := getCredentials(inputs.Driver.Secrets)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Unable to provision the resource",
			errors.New("RES-103", "bad credentials", nil, err))
		return
	}

	// Parse the resource cookie
	//
	var cookie = ResourceCookie{CreatedAt: time.Now().UTC()}
	if err := cookies.Decode(r.Header.Get(cookies.HeaderHumanitecDriverCookie), &cookie); err != nil {
		writeError(w, http.StatusBadRequest, "Unable to parse the resource cookie", err)
		return
	}

	// Prepare outputs (draft)
	//
	res := DriverOutputs{
		GUResID:  GUResID,
		Type:     inputs.Type,
		Resource: cookie.Resource,
	}

	// Provision a new resource
	//
	if cookie.GUResID == "" {
		if data, err := createS3Bucket(ctx, &inputs, region, accessKeyId, secretAccessKey, api.newAwsClient); err != nil {
			writeError(w, http.StatusBadRequest, "Unable to provision the resource", err)
			return
		} else {
			res.Resource = data
		}
	}

	// Refresh AWS credentials
	//
	res.Resource.Secrets["aws_access_key_id"] = accessKeyId
	res.Resource.Secrets["aws_secret_access_key"] = secretAccessKey

	// Set/Update the resource cookie
	//
	cookie = ResourceCookie{
		GUResID:   res.GUResID,
		Type:      res.Type,
		CreatedAt: cookie.CreatedAt,

		Region:          region,
		AWSAccessKeyId:  accessKeyId,
		AWSAccessSecret: secretAccessKey,

		Resource: res.Resource,
	}
	if cookieHdrValue, err := cookies.Encode(&cookie); err != nil {
		writeError(w, http.StatusBadRequest, "Failed to build the resource cookie", err)
		return
	} else {
		w.Header().Add(cookies.HeaderHumanitecDriverCookieSet, cookieHdrValue)
	}

	writeAsJSON(w, http.StatusOK, res)
}

// oas:operation DELETE /s3/{GUResID}; Deletes the resource.
//
// oas:param GUResID; The resource Id.
//
// oas:response 204; Deleted successfully.
// oas:response 404 herrors.HumanitecError; The resource not found.
func (api *apiServer) deleteS3(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	GUResID := params["GUResID"]
	ctx := r.Context()

	if !isValidAsID(GUResID) {
		writeError(w, http.StatusBadRequest, "GUResID is not a valid Humanitec ID", nil)
		return
	}

	// Parse the resource cookie
	//
	var cookie ResourceCookie
	if err := cookies.Decode(r.Header.Get(cookies.HeaderHumanitecDriverCookie), &cookie); err != nil {
		writeError(w, http.StatusBadRequest, "Unable to parse the resource cookie", err)
		return
	} else if cookie.GUResID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if cookie.Type != "s3" {
		writeError(w, http.StatusBadRequest, "Invalid resource type", nil)
		return
	}

	// Delete the resource
	//
	if err := deleteS3Bucket(ctx, &cookie, api.newAwsClient); err != nil {
		writeError(w, http.StatusBadRequest,
			fmt.Sprintf("Unable to delete the S3 bucket record '%s'", cookie.Resource.Values["bucket"]),
			err)
		return
	}

	w.Header().Set(cookies.HeaderHumanitecDriverCookieSet, "")
	w.WriteHeader(http.StatusNoContent)
}
