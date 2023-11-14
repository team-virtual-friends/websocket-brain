package common

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/sieglu2/virtual-friends-brain/foundation"
)

type DatastoreClient struct {
	client *datastore.Client
}

type CharacterInfo struct {
	CharacterId string `datastore:"character_id"`

	AvatarUrl        string `datastore:"rpm_url"`      // it's not really rpm... but fine.
	ElevenLabId      string `datastore:"elevanlab_id"` // typo... but fine.
	Gender           string `datastore:"gender"`
	Name             string `datastore:"name"`
	Greeting         string `datastore:"character_greeting"`
	DescriptionGcsId string `datastore:"character_description"`
	PromptsGcsId     string `datastore:"character_prompts"`
	OpenaiApiKey     string `datastore:"api_key"`
	AssistantId      string `datastore:"assistant_id"`

	// description and prompts need to be looked up in GCS later, not from datastore directly.
	Description string
	Prompts     string
}

func NewDatastoreClient(ctx context.Context) (*DatastoreClient, error) {
	client, err := datastore.NewClient(ctx, GcpProjectId)
	if err != nil {
		return nil, err
	}
	return &DatastoreClient{
		client: client,
	}, nil
}

func (t *DatastoreClient) QueryCharacter(ctx context.Context, characterId string) (*CharacterInfo, error) {
	logger := foundation.Logger()

	query := datastore.NewQuery("Character").
		Namespace("characters_db").
		FilterField("character_id", "=", characterId).
		Limit(1)

	var results []*CharacterInfo
	_, err := t.client.GetAll(ctx, query, &results)
	if err != nil {
		if strings.Contains(err.Error(), "no such struct field") {
			logger.Warnf("QueryCharacter: %+v", err)
		} else {
			err = fmt.Errorf("failed to GetAll from datastore: %v", err)
			logger.Error(err)
			return nil, err
		}
	}
	if len(results) == 0 {
		err = fmt.Errorf("empty result for characterId (%s)", characterId)
		logger.Error(err)
		return nil, err
	}

	return results[0], nil
}
