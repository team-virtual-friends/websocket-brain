package common

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
)

type GcsClient struct {
	client *storage.Client
}

func NewGcsClient(ctx context.Context) (*GcsClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &GcsClient{
		client: client,
	}, nil
}

func (t *GcsClient) DownloadBlob(ctx context.Context, bucketName, path string) ([]byte, error) {
	bucket := t.client.Bucket(bucketName)
	object := bucket.Object(path)
	reader, err := object.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(reader)
}
