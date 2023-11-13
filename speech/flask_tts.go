package speech

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/sieglu2/virtual-friends-brain/foundation"
	"github.com/sieglu2/virtual-friends-brain/virtualfriends_go"
)

func TextToSpeechWithFlask(ctx context.Context, text string, gender virtualfriends_go.Gender) ([]byte, error) {
	logger := foundation.Logger()

	output, err := foundation.AccessLocalFlask(ctx, "text_to_speech", map[string]string{
		"text":   text,
		"gender": gender.String(),
	})
	if err != nil {
		err = fmt.Errorf("error calling AccessLocalFlask for text_to_speech: %v", err)
		logger.Error(err)
		return nil, err
	}

	decodedMp3Data, err := base64.StdEncoding.DecodeString(string(output))
	if err != nil {
		err = fmt.Errorf("error decoding for pitch_shift: %v", err)
		logger.Error(err)
		return nil, err
	}

	wavBytes, err := Mp3ToWav(decodedMp3Data)
	if err != nil {
		err = fmt.Errorf("failed to convert mp3 to wav: %v", err)
		logger.Error(err)
		return nil, err
	}

	switch gender {
	case virtualfriends_go.Gender_Gender_Female:
		wavBytes, err = PitchShift(ctx, wavBytes, 0.1)
		if err != nil {
			err = fmt.Errorf("failed to pitch shift: %v", err)
			logger.Error(err)
			return nil, err
		}
	}

	return wavBytes, nil
}
