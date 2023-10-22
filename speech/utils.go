package speech

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/sieglu2/virtual-friends-brain/foundation"
)

func PitchShift(inputData []byte, pitchShiftFactor float64) ([]byte, error) {
	logger := foundation.Logger()

	encodedData := base64.StdEncoding.EncodeToString(inputData)

	url := "http://localhost:8511/pitch_shift"
	payload := []byte(fmt.Sprintf(`{"octaves": "%f", "b64_encoded": "%s"}`, pitchShiftFactor, encodedData))

	// Create a request with the payload
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		err = fmt.Errorf("error creating request: %v", err)
		logger.Error(err)
		return nil, err
	}

	// Set the content type for the request
	req.Header.Set("Content-Type", "application/json")

	// Create an HTTP client
	client := &http.Client{}

	// Perform the POST request
	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("error sending request: %v", err)
		logger.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	output, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("error reading response: %v", err)
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
