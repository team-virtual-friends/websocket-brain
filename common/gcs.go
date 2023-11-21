package common

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"github.com/sieglu2/virtual-friends-brain/foundation"
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
	logger := foundation.Logger()

	bucket := t.client.Bucket(bucketName)
	object := bucket.Object(path)
	reader, err := object.NewReader(ctx)
	if err != nil {
		err = fmt.Errorf("failed to read from gcs for %s: %v", path, err)
		logger.Error(err)
		return nil, err
	}
	bytes, err := io.ReadAll(reader)
	if err != nil {
		err = fmt.Errorf("failed to read bytes for %s: %v", path, err)
		logger.Error(err)
		return nil, err
	}
	if err := reader.Close(); err != nil {
		err = fmt.Errorf("failed to close reader for gcs file %s/%s: %v", bucketName, path, err)
		logger.Error(err)
		// not return err.
	}
	return bytes, nil
}

func (t *GcsClient) ExtendCharacterInfo(ctx context.Context, character *CharacterInfo) error {
	logger := foundation.Logger()

	if len(character.DescriptionGcsId) > 0 {
		description, err := t.fetchProxyCharacterInfo(ctx, character.CharacterId, "character_description", character.DescriptionGcsId)
		if err != nil {
			err = fmt.Errorf("failed to fetch description from gcs: %v", err)
			logger.Error(err)
			return err
		}
		character.Description = description
	}

	if len(character.PromptsGcsId) > 0 {
		prompts, err := t.fetchProxyCharacterInfo(ctx, character.CharacterId, "character_prompts", character.PromptsGcsId)
		logger.Errorf("character.PromptsGcsId: " + character.PromptsGcsId)
		if err != nil {
			err = fmt.Errorf("failed to fetch prompts from gcs: %v", err)
			logger.Error(err)
			return err
		}
		character.Prompts = prompts
	}

	return nil
}

func (t *GcsClient) fetchProxyCharacterInfo(ctx context.Context, characterId, attributeName, attributeValue string) (string, error) {
	logger := foundation.Logger()

	if len(attributeValue) == 0 {
		err := fmt.Errorf("empty attributeValue")
		logger.Error(err)
		return "", err
	}

	// bucket := t.client.Bucket("datastore_large_data")
	bucketName := "datastore_large_data"
	path := fmt.Sprintf("%s/%s/%s", characterId, attributeName, attributeValue)
	content, err := t.LoadContentFromGcs(ctx, bucketName, path)
	if err != nil {
		err = fmt.Errorf("failed to fetchProxyCharacterInfo for %s: %v", path, err)
		logger.Error(err)
		return "", err
	}

	return content, nil
}

func (t *GcsClient) SaveContentToGcs(ctx context.Context, bucketName, path string, content string) error {
	logger := foundation.Logger()

	bucket := t.client.Bucket(bucketName)
	object := bucket.Object(path)
	writer := object.NewWriter(ctx)
	if _, err := writer.Write([]byte(content)); err != nil {
		err = fmt.Errorf("failed to write to gcs for %s/%s: %v", bucketName, path, err)
		logger.Error(err)
		return err
	}
	if err := writer.Close(); err != nil {
		err = fmt.Errorf("failed to close writer for gcs file %s/%s: %v", bucketName, path, err)
		logger.Error(err)
		// not return err.
	}

	return nil
}

func (t *GcsClient) LoadContentFromGcs(ctx context.Context, bucketName, path string) (string, error) {
	logger := foundation.Logger()

	bytes, err := t.DownloadBlob(ctx, bucketName, path)
	if err != nil {
		err = fmt.Errorf("failed to DownloadBlob for %s/%s: %v", bucketName, path, err)
		logger.Error(err)
		return "", err
	}

	return string(bytes), nil
}
