package common

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"github.com/sieglu2/virtual-friends-brain/foundation"
)

const (
	chatHistoryBucketName = "vf-chat-histories"
)

var (
	ObjectNotExistError = storage.ErrObjectNotExist
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
		if err == storage.ErrObjectNotExist {
			return nil, ObjectNotExistError
		}
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

func (t *GcsClient) LoadChatHistory(ctx context.Context, toCharacterId, fromCharacterId string) (string, error) {
	logger := foundation.Logger()

	// using toCharacterId as the first component so we can query who has talked to me.
	filePath := fmt.Sprintf("%s/%s.txt", toCharacterId, fromCharacterId)
	bytes, err := t.DownloadBlob(ctx, chatHistoryBucketName, filePath)
	if err != nil {
		if err == ObjectNotExistError {
			logger.Infof("did not find %s, treat it empty", filePath)
			return "", nil
		}

		err = fmt.Errorf("failed to DownloadBlob chatHistory for %s: %v", filePath, err)
		logger.Error(err)
		return "", err
	}

	return string(bytes), nil
}

func (t *GcsClient) SaveChatHistory(ctx context.Context, toCharacterId, fromCharacterId string, completeChatHistory string) error {
	logger := foundation.Logger()

	// using toCharacterId as the first component so we can query who has talked to me.
	filePath := fmt.Sprintf("%s/%s.txt", toCharacterId, fromCharacterId)

	bucket := t.client.Bucket(chatHistoryBucketName)
	obj := bucket.Object(filePath)

	w := obj.NewWriter(ctx)
	_, err := w.Write([]byte(completeChatHistory))
	if err != nil {
		err = fmt.Errorf("failed to write chatHistory to gcs(%s): %v", filePath, err)
		logger.Error(err)
		return err
	}
	if err := w.Close(); err != nil {
		err = fmt.Errorf("Failed to close gcs(%s) writer: %v", filePath, err)
		logger.Error(err)
		return err
	}

	return nil
}

func (t *GcsClient) ExtendCharacterInfo(ctx context.Context, character *CharacterInfo) error {
	logger := foundation.Logger()

	if len(character.DescriptionGcsId) > 0 {
		description, err := t.fetchCharacterInfo(ctx, character.CharacterId, "character_description", character.DescriptionGcsId)
		if err != nil {
			err = fmt.Errorf("failed to fetch description from gcs: %v", err)
			logger.Error(err)
			return err
		}
		character.Description = description
	}

	if len(character.PromptsGcsId) > 0 {
		prompts, err := t.fetchCharacterInfo(ctx, character.CharacterId, "character_prompts", character.PromptsGcsId)
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
		if err == storage.ErrObjectNotExist {
			return "", ObjectNotExistError
		}
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
