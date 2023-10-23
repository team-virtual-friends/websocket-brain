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
	return bytes, nil
}

func (t *GcsClient) ExtendCharacterInfo(ctx context.Context, character *CharacterInfo) error {
	logger := foundation.Logger()

	description, err := t.fetchCharacterInfo(ctx, character.CharacterId, "character_description", character.DescriptionGcsId)
	if err != nil {
		err = fmt.Errorf("failed to fetch description from gcs: %v", err)
		logger.Error(err)
		return err
	}
	character.Description = description

	prompts, err := t.fetchCharacterInfo(ctx, character.CharacterId, "character_prompts", character.PromptsGcsId)
	if err != nil {
		err = fmt.Errorf("failed to fetch prompts from gcs: %v", err)
		logger.Error(err)
		return err
	}
	character.Prompts = prompts

	return nil
}

func (t *GcsClient) fetchCharacterInfo(ctx context.Context, characterId, attributeName, attributeValue string) (string, error) {
	logger := foundation.Logger()

	bucket := t.client.Bucket("datastore_large_data")
	if len(attributeValue) == 0 {
		err := fmt.Errorf("empty attributeValue")
		logger.Error(err)
		return "", err
	}

	path := fmt.Sprintf("%s/%s/%s", characterId, attributeName, attributeValue)
	object := bucket.Object(path)
	reader, err := object.NewReader(ctx)
	if err != nil {
		err = fmt.Errorf("failed to read from gcs for %s: %v", path, err)
		logger.Error(err)
		return "", err
	}

	bytes, err := io.ReadAll(reader)
	if err != nil {
		err = fmt.Errorf("failed to read bytes for %s: %v", path, err)
		logger.Error(err)
		return "", err
	}

	return string(bytes), nil
}
