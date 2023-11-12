package common

import "github.com/sashabaranov/go-openai"

const (
	openAiApiKey = "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"
)

func NewOpenAiClient() *openai.Client {
	return openai.NewClient(openAiApiKey)
}
