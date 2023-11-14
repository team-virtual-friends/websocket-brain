package speech

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/iFaceless/godub"
	"github.com/iFaceless/godub/converter"
	"github.com/pemistahl/lingua-go"
	"github.com/sieglu2/virtual-friends-brain/foundation"
)

var (
	linguaLanguageDetector = lingua.NewLanguageDetectorBuilder().FromAllLanguages().Build()
)

func PitchShift(ctx context.Context, wavBytes []byte, pitchShiftFactor float64) ([]byte, error) {
	logger := foundation.Logger()

	encodedData := base64.StdEncoding.EncodeToString(wavBytes)

	output, err := foundation.AccessLocalFlask(ctx, "pitch_shift", "POST", map[string]string{
		"octaves":     fmt.Sprintf("%f", pitchShiftFactor),
		"b64_encoded": encodedData,
	})
	if err != nil {
		err = fmt.Errorf("error calling AccessLocalFlask for pitch_shift: %v", err)
		logger.Error(err)
		return nil, err
	}

	decodedData, err := base64.StdEncoding.DecodeString(string(output))
	if err != nil {
		err = fmt.Errorf("error decoding for pitch_shift: %v", err)
		logger.Error(err)
		return nil, err
	}

	return decodedData, nil
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

func DetectShortLanguageCode(text string) string {
	if language, exists := linguaLanguageDetector.DetectLanguageOf(text); exists {
		return strings.ToLower(language.IsoCode639_1().String())
	}
	return ""
}
