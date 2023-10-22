package speech

import (
	"encoding/base64"
	"fmt"

	"github.com/sieglu2/virtual-friends-brain/foundation"
)

func PitchShift(inputData []byte, pitchShiftFactor float64) ([]byte, error) {
	logger := foundation.Logger()

	encodedData := base64.StdEncoding.EncodeToString(inputData)

	output, err := foundation.AccessLocalFlask("pitch_shift", map[string]string{
		"octaves":     fmt.Sprintf("%f", pitchShiftFactor),
		"b64_encoded": encodedData,
	})
	if err != nil {
		err = fmt.Errorf("error calling AccessLocalFlask: %v", err)
		logger.Error(err)
		return nil, err
	}

	decodedData, err := base64.StdEncoding.DecodeString(string(output))
	if err != nil {
		err = fmt.Errorf("error decoding: %v", err)
		logger.Error(err)
		return nil, err
	}

	return decodedData, nil
}
