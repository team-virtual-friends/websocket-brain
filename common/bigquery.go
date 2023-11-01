package common

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/sieglu2/virtual-friends-brain/foundation"
)

const (
	datasetIdVirtualFriends = "virtualfriends"
	bigqueryWriteTimeout    = time.Second
)

type BigQueryClient struct {
	client *bigquery.Client
}

type LatencyStats struct {
	Env          string    `bigquery:"env"`
	SessionId    string    `bigquery:"session_id"`
	UserId       string    `bigquery:"user_id"`
	UserIp       string    `bigquery:"user_ip"`
	CharacterId  string    `bigquery:"character_id"`
	LatencyType  string    `bigquery:"latency_type"`
	LatencyValue float64   `bigquery:"latency_value"`
	Timestamp    time.Time `bigquery:"timestamp"`
}

type ChatHistory struct {
	UserId        string    `bigquery:"user_id"`
	UserIp        string    `bigquery:"user_ip"`
	CharacterId   string    `bigquery:"character_id"`
	ChatHistory   string    `bigquery:"chat_history"`
	Timestamp     time.Time `bigquery:"timestamp"`
	ChatSessionId string    `bigquery:"chat_session_id"`
	RuntimeEnv    string    `bigquery:"runtime_env"`
}

func NewBigQueryClient(ctx context.Context) (*BigQueryClient, error) {
	client, err := bigquery.NewClient(ctx, GcpProjectId)
	if err != nil {
		return nil, err
	}

	return &BigQueryClient{
		client: client,
	}, nil
}

func (t BigQueryClient) writeData(ctx context.Context, datasetId, tableId string, data interface{}) error {
	logger := foundation.Logger()

	datasetRef := t.client.Dataset(datasetId)
	tableRef := datasetRef.Table(tableId)
	uploader := tableRef.Uploader()

	if err := uploader.Put(ctx, data); err != nil {
		err = fmt.Errorf("failed to uploader.Put to bigquery: %v", err)
		logger.Error(err)
		return err
	}
	return nil
}

func (t *BigQueryClient) WriteLatencyStats(ctx context.Context, latencyStats *LatencyStats) error {
	logger := foundation.Logger()

	err := t.writeData(ctx, datasetIdVirtualFriends, "latency_log", []*LatencyStats{latencyStats})
	if err != nil {
		err = fmt.Errorf("failed to writeData for LatencyStats: %v", err)
		logger.Error(err)
		return err
	}

	return nil
}

func (t *BigQueryClient) WriteChatHistory(ctx context.Context, chatHistory *ChatHistory) error {
	logger := foundation.Logger()

	err := t.writeData(ctx, datasetIdVirtualFriends, "chat_history", []*ChatHistory{chatHistory})
	if err != nil {
		err = fmt.Errorf("failed to writeData for ChatHistory: %v", err)
		logger.Error(err)
		return err
	}

	return nil
}
