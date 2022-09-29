package api

import (
	"context"
	"encoding/json"
	"fmt"
	"humanitec.io/custom-reference-driver/internal/testutils"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"gotest.tools/assert"

	"humanitec.io/custom-reference-driver/internal/aws"
	aws_mock "humanitec.io/custom-reference-driver/internal/aws/mocks"
	"humanitec.io/custom-reference-driver/internal/cookies"
	"humanitec.io/custom-reference-driver/internal/errors"
)

func TestCreateS3Bucket(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accessKeyId := "AWS_ACCESS_KEY_ID-value"
	secretAccessKey := "AWS_SECRET_ACCESS_KEY-value"
	region := "eu-west-1"

	var a = aws_mock.NewMockClient(ctrl)
	var newAwsClient = func(key, secret, reg string) (aws.Client, error) {
		assert.Equal(t, key, accessKeyId)
		assert.Equal(t, secret, secretAccessKey)
		assert.Equal(t, reg, region)
		return a, nil
	}

	inputs := DriverInputs{
		Type:     "s3",
		Resource: map[string]interface{}{},
		Driver: &ValuesSecrets{
			Values: map[string]interface{}{
				"region": region,
			},
			Secrets: map[string]interface{}{
				"account": map[string]interface{}{
					"aws_access_key_id":     accessKeyId,
					"aws_secret_access_key": secretAccessKey,
				},
			},
		},
	}
	expectedData := ValuesSecrets{
		Values: map[string]interface{}{
			"region": region,
		},
		Secrets: map[string]interface{}{},
	}

	a.
		EXPECT().
		CreateBucket(testutils.WithTestContext(), gomock.AssignableToTypeOf("")).
		Do(func(_ context.Context, bn interface{}) {
			expectedData.Values["bucket"] = bn.(string)
		}).
		Return(region, nil).
		Times(1)

	responseData, err := createS3Bucket(testutils.TestContext(), &inputs, region, accessKeyId, secretAccessKey, newAwsClient)

	assert.NilError(t, err, "Unexpected error: %v", err)
	assert.DeepEqual(t, &expectedData, responseData)
}

func TestDeleteS3Bucket(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accessKeyId := "AWS_ACCESS_KEY_ID-value"
	secretAccessKey := "AWS_SECRET_ACCESS_KEY-value"
	region := "eu-west-1"

	var a = aws_mock.NewMockClient(ctrl)
	var newAwsClient = func(key, secret, reg string) (aws.Client, error) {
		assert.Equal(t, key, accessKeyId)
		assert.Equal(t, secret, secretAccessKey)
		assert.Equal(t, reg, region)
		return a, nil
	}

	cookie := ResourceCookie{
		GUResID:         "resource-id",
		Type:            "s3",
		CreatedAt:       time.Date(2020, 07, 16, 18, 12, 20, 0, time.UTC),
		Region:          region,
		AWSAccessKeyId:  accessKeyId,
		AWSAccessSecret: secretAccessKey,
		Resource: &ValuesSecrets{
			Values: map[string]interface{}{
				"region": "http://my-bucket.s3.amazonaws.com/",
				"bucket": "my-bucket",
			},
		},
	}

	a.
		EXPECT().
		DeleteBucket(testutils.WithTestContext(), cookie.Resource.Values["bucket"]).
		Return(nil).
		Times(1)

	err := deleteS3Bucket(testutils.TestContext(), &cookie, newAwsClient)

	assert.NilError(t, err, "Unexpected error: %v", err)
}

func TestUpsertS3_CreateNew(t *testing.T) {
	const (
		GUResID         = "test-db-id"
		resType         = "s3"
		region          = "eu-west-1"
		accessKeyId     = "AWS_ACCESS_KEY_ID-value"
		secretAccessKey = "AWS_SECRET_ACCESS_KEY-value"
	)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := aws_mock.NewMockClient(ctrl)
	newAwsClient := func(key, secret, reg string) (aws.Client, error) {
		assert.Equal(t, key, accessKeyId)
		assert.Equal(t, secret, secretAccessKey)
		assert.Equal(t, reg, region)
		return a, nil
	}

	inputs := DriverInputs{
		Type:     resType,
		Resource: nil,
		Driver: &ValuesSecrets{
			Values: map[string]interface{}{
				"region": region,
			},
			Secrets: map[string]interface{}{
				"account": map[string]interface{}{
					"aws_access_key_id":     accessKeyId,
					"aws_secret_access_key": secretAccessKey,
				},
			},
		},
	}

	var bucket string
	a.
		EXPECT().
		CreateBucket(testutils.WithTestContext(), gomock.AssignableToTypeOf("")).
		Do(func(_ context.Context, bn interface{}) {
			bucket = bn.(string)
		}).
		Return(region, nil).
		Times(1)

	r := mux.NewRouter()
	(&apiServer{
		newAwsClient: newAwsClient,
	}).MapRoutes(r)

	resp := testutils.ExecuteTestRequest(testutils.TestContext(), t, r, http.MethodPut, fmt.Sprintf("/s3/%s", GUResID), nil, inputs)

	// Confirm status code
	//
	expectedCode := http.StatusOK
	assert.Equal(t, resp.Code, expectedCode,
		fmt.Sprintf("Should return HTTP %v. Actual: HTTP %v", expectedCode, resp.Code))

	// Confirm resource cookie
	//
	var resCookie ResourceCookie
	var resCookieHdr = resp.Header().Get(cookies.HeaderHumanitecDriverCookieSet)
	err := cookies.Decode(resCookieHdr, &resCookie)
	assert.NilError(t, err, "Unable to decode cookie header: %s : %v", resCookieHdr, err)
	assert.DeepEqual(t, resCookie, ResourceCookie{
		GUResID:         GUResID,
		Type:            inputs.Type,
		CreatedAt:       resCookie.CreatedAt,
		Region:          region,
		AWSAccessKeyId:  accessKeyId,
		AWSAccessSecret: secretAccessKey,
		Resource: &ValuesSecrets{
			Values: map[string]interface{}{
				"region": region,
				"bucket": bucket,
			},
			Secrets: map[string]interface{}{
				"aws_access_key_id":     accessKeyId,
				"aws_secret_access_key": secretAccessKey,
			},
		},
	})

	// Confirm response data
	//
	var resData DriverOutputs
	err = json.Unmarshal(resp.Body.Bytes(), &resData)
	assert.NilError(t, err, "Unable to parse the response: %s : %v", resp.Body, err)
	assert.DeepEqual(t, resData, DriverOutputs{
		GUResID: GUResID,
		Type:    resType,
		Resource: &ValuesSecrets{
			Values: map[string]interface{}{
				"region": region,
				"bucket": bucket,
			},
			Secrets: map[string]interface{}{
				"aws_access_key_id":     accessKeyId,
				"aws_secret_access_key": secretAccessKey,
			},
		},
	})
}

func TestUpsertS3_Existing(t *testing.T) {
	const (
		GUResID         = "test-db-id"
		resType         = "s3"
		bucket          = "my-s3-bucket"
		region          = "eu-west-1"
		accessKeyId     = "AWS_ACCESS_KEY_ID-value"
		secretAccessKey = "AWS_SECRET_ACCESS_KEY-value"
	)

	cookie := ResourceCookie{
		GUResID:         GUResID,
		Type:            resType,
		CreatedAt:       time.Now().UTC(),
		Region:          region,
		AWSAccessKeyId:  accessKeyId,
		AWSAccessSecret: secretAccessKey,
		Resource: &ValuesSecrets{
			Values: map[string]interface{}{
				"region": region,
				"bucket": bucket,
			},
			Secrets: map[string]interface{}{
				"aws_access_key_id":     accessKeyId,
				"aws_secret_access_key": secretAccessKey,
			},
		},
	}
	cookieHdr, _ := cookies.Encode(&cookie)

	inputs := DriverInputs{
		Type:     resType,
		Resource: nil,
		Driver: &ValuesSecrets{
			Values: map[string]interface{}{
				"region": region,
			},
			Secrets: map[string]interface{}{
				"account": map[string]interface{}{
					"aws_access_key_id":     accessKeyId,
					"aws_secret_access_key": secretAccessKey,
				},
			},
		},
	}

	r := mux.NewRouter()
	(&apiServer{}).MapRoutes(r)

	headers := map[string]string{
		cookies.HeaderHumanitecDriverCookie: cookieHdr,
	}
	resp := testutils.ExecuteTestRequest(context.TODO(), t, r, http.MethodPut, fmt.Sprintf("/s3/%s", GUResID), headers, inputs)

	// Confirm status code
	//
	expectedCode := http.StatusOK
	assert.Equal(t, resp.Code, expectedCode,
		fmt.Sprintf("Should return HTTP %v. Actual: HTTP %v", expectedCode, resp.Code))

	// Confirm resource cookie
	//
	var resCookie ResourceCookie
	var resCookieHdr = resp.Header().Get(cookies.HeaderHumanitecDriverCookieSet)
	err := cookies.Decode(resCookieHdr, &resCookie)
	assert.NilError(t, err, "Unable to decode cookie header: %s : %v", resCookieHdr, err)
	assert.DeepEqual(t, resCookie, ResourceCookie{
		GUResID:         GUResID,
		Type:            inputs.Type,
		CreatedAt:       resCookie.CreatedAt,
		Region:          region,
		AWSAccessKeyId:  accessKeyId,
		AWSAccessSecret: secretAccessKey,
		Resource: &ValuesSecrets{
			Values: map[string]interface{}{
				"region": region,
				"bucket": bucket,
			},
			Secrets: map[string]interface{}{
				"aws_access_key_id":     accessKeyId,
				"aws_secret_access_key": secretAccessKey,
			},
		},
	})

	// Confirm response data
	//
	var resData DriverOutputs
	err = json.Unmarshal(resp.Body.Bytes(), &resData)
	assert.NilError(t, err, "Unable to parse the response: %s : %v", resp.Body, err)
	assert.DeepEqual(t, resData, DriverOutputs{
		GUResID: GUResID,
		Type:    resType,
		Resource: &ValuesSecrets{
			Values: map[string]interface{}{
				"region": region,
				"bucket": bucket,
			},
			Secrets: map[string]interface{}{
				"aws_access_key_id":     accessKeyId,
				"aws_secret_access_key": secretAccessKey,
			},
		},
	})
}

func TestUpsertS3_Error(t *testing.T) {
	const (
		GUResID         = "test-db-id"
		resType         = "s3"
		region          = "eu-west-1"
		accessKeyId     = "AWS_ACCESS_KEY_ID-value"
		secretAccessKey = "AWS_SECRET_ACCESS_KEY-value"
	)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := aws_mock.NewMockClient(ctrl)
	newAwsClient := func(key, secret, reg string) (aws.Client, error) {
		assert.Equal(t, key, accessKeyId)
		assert.Equal(t, secret, secretAccessKey)
		assert.Equal(t, reg, region)
		return a, nil
	}

	inputs := DriverInputs{
		Type:     resType,
		Resource: nil,
		Driver: &ValuesSecrets{
			Values: map[string]interface{}{
				"region": region,
			},
			Secrets: map[string]interface{}{
				"account": map[string]interface{}{
					"aws_access_key_id":     accessKeyId,
					"aws_secret_access_key": secretAccessKey,
				},
			},
		},
	}

	a.
		EXPECT().
		CreateBucket(testutils.WithTestContext(), gomock.AssignableToTypeOf("")).
		Return("", errors.New("DEP-001", "test error", nil, nil)).
		Times(1)

	r := mux.NewRouter()
	(&apiServer{
		newAwsClient: newAwsClient,
	}).MapRoutes(r)

	resp := testutils.ExecuteTestRequest(testutils.TestContext(), t, r, http.MethodPut, fmt.Sprintf("/s3/%s", GUResID), nil, inputs)

	// Confirm status code
	//
	expectedCode := http.StatusBadRequest
	assert.Equal(t, resp.Code, expectedCode,
		fmt.Sprintf("Should return HTTP %v. Actual: HTTP %v", expectedCode, resp.Code))

	// Confirm resource cookie
	//
	var resCookieHdr = resp.Header().Get(cookies.HeaderHumanitecDriverCookieSet)
	assert.Equal(t, resCookieHdr, "")

	// Confirm response data
	//
	var resError *errors.HumanitecError
	err := json.Unmarshal(resp.Body.Bytes(), &resError)
	assert.NilError(t, err, "Unable to parse the response: %s : %v", resp.Body, err)
	assert.Equal(t, resError.Code, "RES-104")
}

func TestDeleteS3(t *testing.T) {
	const (
		GUResID         = "test-db-id"
		resType         = "s3"
		bucket          = "my-s3-bucket"
		region          = "eu-west-1"
		accessKeyId     = "AWS_ACCESS_KEY_ID-value"
		secretAccessKey = "AWS_SECRET_ACCESS_KEY-value"
	)

	cookie := ResourceCookie{
		GUResID:         GUResID,
		Type:            resType,
		CreatedAt:       time.Now().UTC(),
		Region:          region,
		AWSAccessKeyId:  accessKeyId,     //nosec
		AWSAccessSecret: secretAccessKey, //nosec
		Resource: &ValuesSecrets{
			Values: map[string]interface{}{
				"region": region,
				"bucket": bucket,
			},
			Secrets: map[string]interface{}{
				"aws_access_key_id":     accessKeyId,
				"aws_secret_access_key": secretAccessKey,
			},
		},
	}
	cookieHdr, _ := cookies.Encode(&cookie)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	a := aws_mock.NewMockClient(ctrl)
	newAwsClient := func(key, secret, reg string) (aws.Client, error) {
		assert.Equal(t, key, accessKeyId)
		assert.Equal(t, secret, secretAccessKey)
		assert.Equal(t, reg, region)
		return a, nil
	}

	a.
		EXPECT().
		DeleteBucket(testutils.WithTestContext(), bucket).
		Return(nil).
		Times(1)

	r := mux.NewRouter()
	(&apiServer{
		newAwsClient: newAwsClient,
	}).MapRoutes(r)

	headers := map[string]string{
		cookies.HeaderHumanitecDriverCookie: cookieHdr,
	}
	resp := testutils.ExecuteTestRequest(testutils.TestContext(), t, r, http.MethodDelete, fmt.Sprintf("/s3/%s", GUResID), headers, nil)

	// Confirm status code
	//
	expectedCode := http.StatusNoContent
	assert.Equal(t, resp.Code, expectedCode,
		fmt.Sprintf("Should return HTTP %v. Actual: HTTP %v", expectedCode, resp.Code))

	// Confirm resource cookie
	//
	var resCookieHdr = resp.Header().Get(cookies.HeaderHumanitecDriverCookieSet)
	assert.Equal(t, resCookieHdr, "")
}

func TestDeleteS3_NotFound(t *testing.T) {
	const (
		GUResID = "test-db-id"
	)

	r := mux.NewRouter()
	(&apiServer{}).MapRoutes(r)
	resp := testutils.ExecuteTestRequest(context.TODO(), t, r, http.MethodDelete, fmt.Sprintf("/s3/%s", GUResID), nil, nil)

	// Confirm status code
	//
	expectedCode := http.StatusNotFound
	assert.Equal(t, resp.Code, expectedCode,
		fmt.Sprintf("Should return HTTP %v. Actual: HTTP %v", expectedCode, resp.Code))

	// Confirm resource cookie
	//
	var resCookieHdr = resp.Header().Get(cookies.HeaderHumanitecDriverCookieSet)
	assert.Equal(t, resCookieHdr, "")
}
