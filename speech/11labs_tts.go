package speech

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/iFaceless/godub"
	"github.com/iFaceless/godub/converter"
	"github.com/sieglu2/virtual-friends-brain/foundation"
)

const (
	apiKey           = "4fb91ffd3e3e3cd35cbf2d19a64fd4e9"
	urlTemplate      = "https://api.elevenlabs.io/v1/text-to-speech/%s?optimize_streaming_latency=3"
	jsonBodyTemplate = `{"text":"%s","model_id":"eleven_monolingual_v1","voice_settings":{"stability":0.9,"similarity_boost":0.9}}`
)

func TextToSpeechWith11Labs(ctx context.Context, text, voiceId string) ([]byte, error) {
	logger := foundation.Logger()

	jsonBody := fmt.Sprintf(jsonBodyTemplate, text)

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf(urlTemplate, voiceId), bytes.NewBuffer([]byte(jsonBody)))
	if err != nil {
		err = fmt.Errorf("failed to create POST request to 11labs: %v", err)
		logger.Error(err)
		return nil, err
	}

	req.Header.Set("Accept", "audio/mpeg")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", apiKey)

	client := &http.Client{}

	// Perform the POST request
	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("error sending request to 11labs: %v", err)
		logger.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	output, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("error reading response from 11labs: %v", err)
		logger.Error(err)
		return nil, err
	}

	logger.Infof("mp3 bytes length: %d", len(output))

	wavBytes, err := Mp3ToWav(output)
	if err != nil {
		err = fmt.Errorf("error converting mp3 to wav: %v", err)
		logger.Error(err)
		return nil, err
	}

	return wavBytes, nil
}

func Mp3ToWav(mp3Data []byte) ([]byte, error) {
	logger := foundation.Logger()
	segment, err := godub.NewLoader().Load(bytes.NewReader(mp3Data))
	if err != nil {
		err = fmt.Errorf("failed to load mp3: %v", err)
		logger.Error(err)
		return nil, err
	}

	wavByteBuffer := bytes.Buffer{}
	err = converter.NewConverter(&wavByteBuffer).
		WithBitRate(int(segment.AsWaveAudio().BitsPerSample)).
		WithDstFormat("wav").
		WithChannels(int(segment.Channels())).
		WithSampleRate(int(segment.AsWaveAudio().SampleRate)).
		Convert(bytes.NewReader(mp3Data))
	if err != nil {
		err = fmt.Errorf("failed to convert mp3 to wav: %v", err)
		logger.Error(err)
		return nil, err
	}
	return wavByteBuffer.Bytes(), nil
}
