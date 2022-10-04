package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

func upsertS3(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	GUResID := params["GUResID"]
	ctx := r.Context()

	if !isValidAsID(GUResID) {
		writeAsJSON(w, http.StatusBadRequest, "GUResID is not a valid Humanitec ID")
		return
	}

	// Parse and validate request payload
	//
	var inputs DriverInputs
	if err := readAsJSON(r.Body, &inputs); err != nil {
		writeAsJSON(w, http.StatusUnprocessableEntity, "Unable to process the request")
		return
	} else if inputs.Type != "s3" {
		writeAsJSON(w, http.StatusBadRequest, "Invalid resource type")
		return
	} else if inputs.Driver == nil {
		inputs.Driver = &ValuesSecrets{}
	}

	// Validate AWS access credentials
	//
	region := inputs.Driver.Values["region"].(string)
	accessKeyId, secretAccessKey, err := getCredentials(inputs.Driver.Secrets)
	if err != nil {
		writeAsJSON(w, http.StatusBadRequest, "Unable to provision the resource")
		return
	}

	// Parse the resource cookie
	//
	var cookie = ResourceCookie{CreatedAt: time.Now().UTC()}
	if err := cookiesDecode(r.Header.Get(HeaderHumanitecDriverCookie), &cookie); err != nil {
		writeAsJSON(w, http.StatusBadRequest, "Unable to parse the resource cookie")
		return
	}

	// Prepare outputs (draft)
	//
	res := DriverOutputs{
		GUResID:  GUResID,
		Type:     inputs.Type,
		Resource: cookie.Resource,
	}

	bucketNameUUID, err := uuid.NewRandom()
	if err != nil {
		writeAsJSON(w, http.StatusInternalServerError, "Unable to generate UUID")
	}
	bucketName := bucketNameUUID.String()

	// Provision a new resource
	//
	if cookie.GUResID == "" {
		if region, err := createBucket(ctx, accessKeyId, secretAccessKey, region, bucketName); err != nil {
			writeAsJSON(w, http.StatusBadRequest, "Unable to provision the resource")
			return
		} else {
			res.Resource = &ValuesSecrets{
				Values: map[string]interface{}{
					"region": region,
					"bucket": bucketName,
				},
				Secrets: map[string]interface{}{},
			}
		}
	} // Don't need to handle existing resource for S3 as it's immediately ready

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
	if cookieHdrValue, err := cookiesEncode(&cookie); err != nil {
		writeAsJSON(w, http.StatusBadRequest, "Failed to build the resource cookie")
		return
	} else {
		w.Header().Add(HeaderHumanitecDriverCookieSet, cookieHdrValue)
	}

	writeAsJSON(w, http.StatusOK, res)
}

func deleteS3(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	GUResID := params["GUResID"]
	ctx := r.Context()

	if !isValidAsID(GUResID) {
		writeAsJSON(w, http.StatusBadRequest, "GUResID is not a valid Humanitec ID")
		return
	}

	// Parse the resource cookie
	//
	var bucketName interface{}
	var cookie ResourceCookie
	if err := cookiesDecode(r.Header.Get(HeaderHumanitecDriverCookie), &cookie); err != nil {
		writeAsJSON(w, http.StatusBadRequest, "Unable to parse the resource cookie")
		return
	} else if cookie.GUResID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if cookie.Type != "s3" {
		writeAsJSON(w, http.StatusBadRequest, "Invalid resource type")
		return
	} else if bucketName = cookie.Resource.Values["bucket"]; bucketName == nil {
		writeAsJSON(w, http.StatusBadRequest, "Missing bucket name")
		return
	}

	// Delete the resource
	//
	if err := deleteBucket(ctx, cookie.AWSAccessKeyId, cookie.AWSAccessSecret, cookie.Region, bucketName.(string)); err != nil {
		writeAsJSON(w, http.StatusBadRequest,
			fmt.Sprintf("Unable to delete the S3 bucket record '%s'", bucketName.(string)))
		return
	}

	w.Header().Set(HeaderHumanitecDriverCookieSet, "")
	w.WriteHeader(http.StatusNoContent)
}

func apiSpec(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./openapi/spec.json")
}

func isAlive(w http.ResponseWriter, _ *http.Request) {
	writeAsText(w, http.StatusOK, fmt.Sprintf("%s %s", AppName, "0.0.0"))
}

func isReady(w http.ResponseWriter, _ *http.Request) {
	writeAsJSON(w, http.StatusOK, map[string]string{
		"app":     AppName,
		"version": "0.0.0",
		"status":  "OK",
	})
}
