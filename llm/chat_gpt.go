package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sieglu2/virtual-friends-brain/foundation"
)

const (
	InferSentimentAndActionPrompt = `
---------
In the end of sentence, based on the conversation, always infer action like [dance], [jump] and  sentiment from the conversation like <happy>, <sad>...
The action should be one of no_action, dance, get_angry, laugh, clap, charm, make_heart, surprise, blow_kiss, backflip, cry, jump, spin.
The sentiment should be one of happy, neutral, sad, angry.
Add the action and sentiment before line separators like .,!?
e.g.
Q: can you show me a dance?
A: sure, I am glad to do it [dance] <happy>.
`

	openAiApiKey = "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"
)

var (
	regexActionAndSentiment = regexp.MustCompile(`^(.*?)\s*(?:\[(.*?)\])?\s*(?:<([^>]*)>)`)
)

type ChatGptClient struct {
	client *openai.Client
}

func NewChatGptClient() *ChatGptClient {
	return &ChatGptClient{
		client: openai.NewClient(openAiApiKey),
	}
}

func (t *ChatGptClient) StreamReplyMessage(
	ctx context.Context, jsonMessages []string,
	process func(replyText string, index int) error,
	completion func() error,
) error {
	logger := foundation.Logger()

	messages := make([]openai.ChatCompletionMessage, 0, len(jsonMessages))
	for _, jsonMessage := range jsonMessages {
		var chatCompletionMessage openai.ChatCompletionMessage
		err := json.Unmarshal([]byte(jsonMessage), &chatCompletionMessage)
		if err != nil {
			err = fmt.Errorf("failed to json.Unmarshal(%s): %v", jsonMessage, err)
			logger.Error(err)
			return err
		}
		if len(chatCompletionMessage.Content) > 0 {
			messages = append(messages, chatCompletionMessage)
		}
	}

	request := openai.ChatCompletionRequest{
		Model:     openai.GPT4,
		MaxTokens: 100,
		Messages:  messages,
		Stream:    true,
	}

	stream, err := t.client.CreateChatCompletionStream(ctx, request)
	if err != nil {
		err = fmt.Errorf("failed to CreateChatCompletionStream: %v", err)
		logger.Error(err)
		return err
	}
	defer stream.Close()

	index := 0
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			logger.Info("gpt reply stream finished")
			if err = completion(); err != nil {
				err = fmt.Errorf("gpt reply completion error: %v", err)
				logger.Error(err)
				return err
			}

			break
		}

		if err != nil {
			err = fmt.Errorf("gpt reply stream error: %v", err)
			logger.Error(err)
			return err
		}

		replyText := response.Choices[0].Delta.Content
		// logger.Infof("receiving stream reply from gpt: %s", replyText)

		if err = process(replyText, index); err != nil {
			err = fmt.Errorf("gpt reply stream process error: %v", err)
			logger.Error(err)
			return err
		}
		index += 1
	}

	return nil
}

// ExtractActionAndSentiment extracts the (reply, action, sentiment string)
// from the string formatted as "reply_text [action] <sentiment>"
func ExtractActionAndSentiment(text string) (string, string, string) {
	// Find the matches
	matches := regexActionAndSentiment.FindStringSubmatch(text)

	if len(matches) == 4 {
		rawText := matches[1]
		action := matches[2]
		sentiment := matches[3]

		if len(strings.Trim(action, " ")) == 0 {
			action = "no_action"
		}
		if len(strings.Trim(sentiment, " ")) == 0 {
			sentiment = "neutral"
		}

		return rawText, action, sentiment
	}

	return text, "", ""
}
