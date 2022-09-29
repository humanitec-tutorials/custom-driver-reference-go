//go:generate mockgen -destination mocks/client.go -package mocks humanitec.io/custom-reference-driver/internal/aws Client

package aws

import (
	"context"
	"humanitec.io/custom-reference-driver/internal/errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Client interface {
	CreateBucket(ctx context.Context, bucketName string) (string, error)
	DeleteBucket(ctx context.Context, bucketName string) error
}

type awsClient struct {
	sess   *session.Session
	region string
}

func New(accessKeyId, secretAccessKey, region string) (Client, error) {
	creds := credentials.NewStaticCredentials(accessKeyId, secretAccessKey, "")
	config := aws.Config{
		Credentials: creds,
	}
	if region != "" {
		config.Region = &region
	}
	sess, err := session.NewSession(&config)
	if err != nil {
		log.Printf(`Error creating AWS Session: %v`, err)
		return nil, errors.New("RES-102", "creating aws session", nil, err)
	}

	return awsClient{
		sess:   sess,
		region: region,
	}, nil
}
