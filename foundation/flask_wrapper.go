package foundation

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	baseUrl = "http://localhost:8511/"
)

func AccessLocalFlask(endpoint string, parameters map[string]string) (string, error) {
	logger := Logger()

	url := baseUrl + endpoint
	paramsBuilder := strings.Builder{}

	paramsBuilder.WriteString("{")
	isFirst := true
	for k, v := range parameters {
		if !isFirst {
			paramsBuilder.WriteString(",")
		} else {
			isFirst = false
		}
		paramsBuilder.WriteString("\"" + k + "\":")
		paramsBuilder.WriteString("\"" + v + "\"")
	}
	paramsBuilder.WriteString("}")

	// Create a request with the payload
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(paramsBuilder.String())))
	if err != nil {
		err = fmt.Errorf("error creating request: %v", err)
		logger.Error(err)
		return "", err
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
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	output, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("error reading response: %v", err)
		logger.Error(err)
		return "", err
	}

	return string(output), nil
}
