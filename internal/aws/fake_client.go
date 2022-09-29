package aws

import "context"

type fakeClient struct {
	region string
}

func (c fakeClient) CreateBucket(ctx context.Context, bucketName string) (string, error) {
	return c.region, nil
}

func (c fakeClient) DeleteBucket(ctx context.Context, bucketName string) error {
	return nil
}

func FakeNew(accessKeyId, secretAccessKey, region string) (Client, error) {
	return fakeClient{
		region: region,
	}, nil
}
