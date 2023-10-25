package speech

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/sieglu2/virtual-friends-brain/foundation"
)

func SpeechToText(ctx context.Context, wavBytes []byte) (string, error) {
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
