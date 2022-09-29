package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func (c awsClient) CreateBucket(ctx context.Context, bucketName string) (string, error) {
	svc := s3.New(c.sess)

	createBucketInput := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(c.region),
		},
	}
	if _, err := svc.CreateBucketWithContext(ctx, createBucketInput); err != nil {
		var aerr awserr.Error
		if errors.As(err, &aerr) {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				return "", fmt.Errorf(`s3 bucket name already exists "%s": %w`, bucketName, aerr)
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				return "", fmt.Errorf(`s3 bucket name already exists "%s": %w`, bucketName, aerr)
			}
		}
		return "", fmt.Errorf(`creating s3 bucket "%s": %w`, bucketName, err)
	}

	headBucketInput := &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	}
	if err := svc.WaitUntilBucketExists(headBucketInput); err != nil {
		return "", fmt.Errorf(`waiting for s3 bucket "%s" to be provisioned: %w`, bucketName, err)
	}

	return c.region, nil
}

func (c awsClient) DeleteBucket(ctx context.Context, bucketName string) error {
	svc := s3.New(c.sess)

	// Empty bucket, before it can be deleted.
	iter := s3manager.NewDeleteListIterator(svc, &s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	})
	if err := s3manager.NewBatchDeleteWithClient(svc).Delete(ctx, iter); err != nil {
		return fmt.Errorf(`deleting objects from s3 bucket "%s": %w`, bucketName, err)
	}

	input := &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	}
	_, err := svc.DeleteBucketWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf(`deleting s3 bucket "%s": %w`, bucketName, err)
	}
	return nil
}
