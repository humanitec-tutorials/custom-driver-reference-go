package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type BucketStatus uint8

const (
	BucketCreated  BucketStatus = 0
	BucketFound    BucketStatus = 1
	BucketNotFound BucketStatus = 2
	BucketError    BucketStatus = 255
)

func getS3(accessKeyId string, secretAccessKey string, region string) (*s3.S3, error) {
	creds := credentials.NewStaticCredentials(accessKeyId, secretAccessKey, "")
	config := aws.Config{
		Credentials: creds,
		Region:      &region,
	}
	sess, err := session.NewSession(&config)
	if err != nil {
		return nil, err
	}
	svc := s3.New(sess)
	return svc, nil
}

func createBucket(ctx context.Context, svc *s3.S3, region string, bucketName string) (string, error) {
	createBucketInput := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(region),
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

	return region, nil
}

func headBucket(ctx context.Context, svc *s3.S3, bucketName string) (BucketStatus, error) {
	headBucketInput := &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	}
	if _, err := svc.HeadBucketWithContext(ctx, headBucketInput); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				return BucketNotFound, nil
			default:
				return BucketError, err
			}
		} else {
			return BucketError, err
		}
	} else {
		return BucketFound, nil
	}
}

func deleteBucket(ctx context.Context, svc *s3.S3, bucketName string) error {

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
