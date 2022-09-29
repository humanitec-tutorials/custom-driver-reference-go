package api

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"humanitec.io/custom-reference-driver/internal/aws"
	"humanitec.io/custom-reference-driver/internal/errors"
)

func createS3Bucket(
	ctx context.Context,
	inputs *DriverInputs,
	region, accessKeyId, secretAccessKey string,
	newAwsClient func(string, string, string) (aws.Client, error),
) (*ValuesSecrets, error) {
	bucketNameUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("creating S3 bucket: generating name: %w", err)
	}
	bucketName := bucketNameUUID.String()

	var res *ValuesSecrets
	if client, err := newAwsClient(accessKeyId, secretAccessKey, region); err != nil {
		return nil, err
	} else if generatedRegion, err := client.CreateBucket(ctx, bucketName); err != nil {
		return nil, errors.New("RES-104", "creating S3 bucket",
			map[string]interface{}{"resource": inputs.Resource, "driver.values": inputs.Driver.Values},
			err)
	} else {
		res = &ValuesSecrets{
			Values: map[string]interface{}{
				"region": generatedRegion,
				"bucket": bucketName,
			},
			Secrets: map[string]interface{}{},
		}
	}

	return res, err
}

func deleteS3Bucket(
	ctx context.Context,
	cookie *ResourceCookie,
	newAwsClient func(string, string, string) (aws.Client, error),
) error {
	bucket := cookie.Resource.Values["bucket"].(string)

	if client, err := newAwsClient(cookie.AWSAccessKeyId, cookie.AWSAccessSecret, cookie.Region); err != nil {
		return err
	} else if err := client.DeleteBucket(ctx, bucket); err != nil {
		return errors.New("RES-104", fmt.Sprintf("deleting S3 bucket record '%s'", bucket),
			map[string]interface{}{"resource.values": cookie.Resource.Values},
			err)
	}

	return nil
}
