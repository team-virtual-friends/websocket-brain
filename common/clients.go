package common

import (
	"context"

	"github.com/sieglu2/virtual-friends-brain/llm"
	"github.com/sieglu2/virtual-friends-brain/speech"
	"golang.org/x/sync/errgroup"
)

// Clients holds all the connections to the upstream
// DO NOT MODIFY it during runtime but only in InitializeClients.
type Clients struct {
	googleTtsClient *speech.GoogleTtsClient

	gcsClient       *GcsClient
	datastoreClient *DatastoreClient
	bigqueryClient  *BigQueryClient

	chatGptClient *llm.ChatGptClient
	whisperClient *speech.WhisperClient
}

var clients *Clients

func InitializeClients(ctx context.Context) error {
	errGroup, groupCtx := errgroup.WithContext(ctx)
	clients = &Clients{}

	errGroup.Go(func() error {
		googleTts, err := speech.NewGoogleTtsClient(groupCtx)
		if err != nil {
			return nil
		}
		clients.googleTtsClient = googleTts
		return nil
	})

	errGroup.Go(func() error {
		gcsClient, err := NewGcsClient(groupCtx)
		if err != nil {
			return err
		}
		clients.gcsClient = gcsClient
		return nil
	})

	errGroup.Go(func() error {
		datastoreClient, err := NewDatastoreClient(groupCtx)
		if err != nil {
			return err
		}
		clients.datastoreClient = datastoreClient
		return nil
	})

	errGroup.Go(func() error {
		bigqueryClient, err := NewBigQueryClient(groupCtx)
		if err != nil {
			return err
		}
		clients.bigqueryClient = bigqueryClient
		return nil
	})

	errGroup.Go(func() error {
		openaiClient := NewOpenAiClient()

		clients.chatGptClient = llm.NewChatGptClient(openaiClient)
		clients.whisperClient = speech.NewWhisperClient(openaiClient)

		return nil
	})

	// wait for all the parallel initialization to finish.
	if err := errGroup.Wait(); err != nil {
		return err
	}

	return nil
}

func GetGlobalClients() *Clients {
	return clients
}

func (t *Clients) GetGoogleTtsClient() *speech.GoogleTtsClient {
	return t.googleTtsClient
}

func (t *Clients) GetGcsClient() *GcsClient {
	return t.gcsClient
}

func (t *Clients) GetDatastoreClient() *DatastoreClient {
	return t.datastoreClient
}

func (t *Clients) GetBigQueryClient() *BigQueryClient {
	return t.bigqueryClient
}

func (t *Clients) GetChatGptClient() *llm.ChatGptClient {
	return t.chatGptClient
}

func (t *Clients) GetWhisperClient() *speech.WhisperClient {
	return t.whisperClient
}
