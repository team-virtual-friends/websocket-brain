package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sieglu2/virtual-friends-brain/foundation"
)

// CreateThreadWithFlask calls the Flask endpoint to create a new thread
func CreateThreadWithFlask(ctx context.Context) (string, error) {
	logger := foundation.Logger()

	output, err := foundation.AccessLocalFlask(ctx, "create_thread", "GET", nil)

	if err != nil {
		err = fmt.Errorf("error calling AccessLocalFlask for create_thread: %v", err)
		logger.Error(err)
		return "", err
	}

	var result map[string]string
	err = json.Unmarshal([]byte(output), &result) // Convert string to []byte
	if err != nil {
		err = fmt.Errorf("error unmarshaling response for create_thread: %v", err)
		logger.Error(err)
		return "", err
	}

	threadID, ok := result["thread_id"]
	if !ok {
		err = fmt.Errorf("thread_id not found in response")
		logger.Error(err)
		return "", err
	}

	return threadID, nil
}

// CreateMessageAndRunThreadWithFlask calls the Flask endpoint to create a message and run a thread
func CreateMessageAndRunThreadWithFlask(ctx context.Context, threadID, apiKey, assistantID, content string) (string, error) {
	logger := foundation.Logger()

	data := map[string]string{
		"thread_id":    threadID,
		"api_key":      apiKey,
		"assistant_id": assistantID,
		"content":      content,
	}

	output, err := foundation.AccessLocalFlask(ctx, "create_message_and_run_thread", "POST", data)
	if err != nil {
		err = fmt.Errorf("error calling AccessLocalFlask for create_message_and_run_thread: %v", err)
		logger.Error(err)
		return "", err
	}

	var result map[string]string
	err = json.Unmarshal([]byte(output), &result) // Convert string to []byte
	if err != nil {
		err = fmt.Errorf("error unmarshaling response for create_message_and_run_thread: %v", err)
		logger.Error(err)
		return "", err
	}

	messageResponse, ok := result["response"]
	if !ok {
		err = fmt.Errorf("response not found in response")
		logger.Error(err)
		return "", err
	}

	return messageResponse, nil
}
