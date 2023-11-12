package speech

import (
	"bytes"
	"context"
	"encoding/base64"
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

// SpeechToText is not working, it complaints format is not supported...
func (t *WhisperClient) SpeechToText(ctx context.Context, wavBytes []byte) (string, error) {
	logger := foundation.Logger()

	request := openai.AudioRequest{
		Model:  "whisper-1",
		Reader: bytes.NewReader(wavBytes),
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

func SpeechToTextViaFlask(ctx context.Context, wavBytes []byte) (string, error) {
	logger := foundation.Logger()

	encodedData := base64.StdEncoding.EncodeToString(wavBytes)
	output, err := foundation.AccessLocalFlask(ctx, "speech_to_text", map[string]string{
		"b64_encoded": encodedData,
	})
	if err != nil {
		err = fmt.Errorf("error calling AccessLocalFlask for speech_to_text: %v", err)
		logger.Error(err)
		return "", err
	}

	return string(output), nil
}
