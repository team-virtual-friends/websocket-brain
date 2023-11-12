package speech

import (
	"bytes"
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/sieglu2/virtual-friends-brain/foundation"
)

type WhisperClient struct {
	client *openai.Client
}

func NewWhisperClient(openaiClient *openai.Client) *WhisperClient {
	return &WhisperClient{
		client: openaiClient,
	}
}

func (t *WhisperClient) SpeechToText(ctx context.Context, wavBytes []byte) (string, error) {
	logger := foundation.Logger()

	request := openai.AudioRequest{
		Model:  "whisper-1",
		Reader: bytes.NewReader(wavBytes),
		// this is needed to tell what format we have for the reader.
		FilePath: "audio.wav",
	}

	response, err := t.client.CreateTranscription(ctx, request)
	if err != nil {
		err = fmt.Errorf("failed to call openai Whisper for speech_to_text: %v", err)
		logger.Error(err)
		return "", err
	}

	logger.Debugf("transcribed text: %s", response.Text)
	return response.Text, nil
}
