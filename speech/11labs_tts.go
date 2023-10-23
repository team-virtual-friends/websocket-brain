package speech

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/sieglu2/virtual-friends-brain/foundation"
)

const (
	apiKey           = "4fb91ffd3e3e3cd35cbf2d19a64fd4e9"
	urlTemplate      = "https://api.elevenlabs.io/v1/text-to-speech/%s?optimize_streaming_latency=3"
	jsonBodyTemplate = `{"text":"%s","model_id":"eleven_monolingual_v1","voice_settings":{"stability":0.9,"similarity_boost":0.9}}`
)

func TextToSpeechWith11Labs(ctx context.Context, text, voiceId string) ([]byte, error) {
	logger := foundation.Logger()

	output, err := foundation.AccessLocalFlask(ctx, "11labs_clone", map[string]string{
		"text":     text,
		"voice_id": voiceId,
	})
	if err != nil {
		err = fmt.Errorf("error calling AccessLocalFlask for 11labs_clone: %v", err)
		logger.Error(err)
		return nil, err
	}

	decodedData, err := base64.StdEncoding.DecodeString(string(output))
	if err != nil {
		err = fmt.Errorf("error decoding for 11labs_clone: %v", err)
		logger.Error(err)
		return nil, err
	}
	return decodedData, nil
}
