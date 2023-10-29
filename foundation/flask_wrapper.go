package foundation

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

type LocalFlaskClient struct {
	httpClient *http.Client

	flaskBaseUrl string
}

var (
	flaskInitOnce = sync.Once{}

	localFlaskClient *LocalFlaskClient
)

func init() {
	flaskPortString := os.Getenv("FLASK_PORT")
	if len(flaskPortString) == 0 {
		flaskPortString = "8085"
	}
	localFlaskClient = &LocalFlaskClient{
		httpClient:   &http.Client{},
		flaskBaseUrl: "http://localhost:" + flaskPortString,
	}
}

func (t *LocalFlaskClient) getUrl(endpoint string) string {
	return t.flaskBaseUrl + "/" + endpoint
}

func (t *LocalFlaskClient) getHttpClient() *http.Client {
	return t.httpClient
}

func AccessLocalFlask(ctx context.Context, endpoint string, parameters map[string]string) (string, error) {
	logger := Logger()

	url := localFlaskClient.getUrl(endpoint)
	logger.Infof("flask url: %s", url)
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
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer([]byte(paramsBuilder.String())))
	if err != nil {
		err = fmt.Errorf("error creating request: %v", err)
		logger.Error(err)
		return "", err
	}

	// Set the content type for the request
	req.Header.Set("Content-Type", "application/json")

	// Perform the POST request
	resp, err := localFlaskClient.getHttpClient().Do(req)
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

	if resp.StatusCode != 200 {
		err = fmt.Errorf("non-200 response status code: %d, error: %s", resp.StatusCode, string(output))
		logger.Error(err)
		return "", err
	}

	return string(output), nil
}
