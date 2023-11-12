package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sieglu2/virtual-friends-brain/common"
	"github.com/sieglu2/virtual-friends-brain/foundation"
	"github.com/sieglu2/virtual-friends-brain/speech"
	"github.com/sieglu2/virtual-friends-brain/virtualfriends_go"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func GenerateVoice(ctx context.Context, vfContext *VfContext, text string, voiceConfig *virtualfriends_go.VoiceConfig) ([]byte, error) {
	logger := foundation.Logger()

	if len(voiceConfig.ElevenLabId) > 0 {
		wavBytes, err := speech.TextToSpeechWith11Labs(ctx, text, voiceConfig.ElevenLabId)
		if err != nil {
			err = fmt.Errorf("failed to TextToSpeechWith11Labs: %v", err)
			logger.Error(err)
			return nil, err
		}
		return wavBytes, nil
	}

	var voiceBytes []byte
	var err error
	switch voiceConfig.VoiceType {
	case virtualfriends_go.VoiceType_VoiceType_NormalMale:
		voiceBytes, err = vfContext.clients.GetGoogleTtsClient().TextToSpeech(ctx, text, virtualfriends_go.Gender_Gender_Male, false)
	case virtualfriends_go.VoiceType_VoiceType_NormalFemale1:
		voiceBytes, err = vfContext.clients.GetGoogleTtsClient().TextToSpeech(ctx, text, virtualfriends_go.Gender_Gender_Female, false)
	case virtualfriends_go.VoiceType_VoiceType_NormalFemale2:
		voiceBytes, err = vfContext.clients.GetGoogleTtsClient().TextToSpeech(ctx, text, virtualfriends_go.Gender_Gender_Female, true)
	case virtualfriends_go.VoiceType_VoiceType_Orc:
		voiceBytes, err = vfContext.clients.GetGoogleTtsClient().TextToSpeech(ctx, text, virtualfriends_go.Gender_Gender_Male, false)
	default:
		return nil, fmt.Errorf("invalid voice_type: %v", voiceConfig.VoiceType)
	}
	if err != nil {
		err = fmt.Errorf("failed to GoogleTtsClient.TextToSpeech: %v", err)
		logger.Error(err)
		return nil, err
	}

	if voiceConfig.Octaves == 0 {
		return voiceBytes, nil
	}
	wavBytes, err := speech.PitchShift(ctx, voiceBytes, float64(voiceConfig.Octaves))
	if err != nil {
		err = fmt.Errorf("failed to PitchShift: %v", err)
		logger.Error(err)
		return nil, err
	}
	return wavBytes, nil
}

func logChatHistory(vfContext *VfContext, chatHistory *common.ChatHistory, eventTime time.Time) error {
	logger := foundation.Logger()

	logger.Infof("logging chat history: %+v", chatHistory)

	bqClient := vfContext.clients.GetBigQueryClient()
	err := bqClient.WriteChatHistory(chatHistory)
	if err != nil {
		err = fmt.Errorf("failed to WriteChatHistory: %v", err)
		logger.Error(err)
		return err
	}

	logger.Info("done writing the chat history")
	return nil
}

func assembleChatHistory(jsonMessages []string) string {
	resultBuilder := strings.Builder{}

	currentRole := ""
	combinedContent := strings.Builder{}
	for _, jsonMessage := range jsonMessages {
		if len(jsonMessage) == 0 {
			continue
		}

		var message ChatMessage
		err := json.Unmarshal([]byte(jsonMessage), &message)
		if err != nil {
			resultBuilder.WriteString("---<missing>---\n")
			continue
		}

		if len(currentRole) > 0 && currentRole != message.Role {
			if currentRole == "assistant" {
				resultBuilder.WriteString("A")
			} else {
				resultBuilder.WriteString(currentRole)
			}

			resultBuilder.WriteString(": ")
			resultBuilder.WriteString(strings.Trim(combinedContent.String(), " "))
			combinedContent.Reset()
		}
		separater := ""
		if combinedContent.Len() > 0 {
			separater = ". "
		}
		combinedContent.WriteString(separater)
		combinedContent.WriteString(message.Content)
		combinedContent.WriteString("\n")
		currentRole = message.Role
	}

	if combinedContent.Len() > 0 {
		if currentRole == "assistant" {
			resultBuilder.WriteString("A")
		} else {
			resultBuilder.WriteString(currentRole)
		}

		resultBuilder.WriteString(": ")
		resultBuilder.WriteString(combinedContent.String())
	}

	return resultBuilder.String()
}

func SpeechToTextViaWhisper(ctx context.Context, client *speech.WhisperClient, wavBytes []byte) (string, error) {
	logger := foundation.Logger()

	backoffConfig := backoff.WithMaxRetries(backoff.NewConstantBackOff(100*time.Millisecond), 3)

	var text string
	err := backoff.Retry(func() error {
		if ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled {
			return nil
		}

		var sttErr error
		timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 500*time.Millisecond)
		text, sttErr = client.SpeechToText(timeoutCtx, wavBytes)
		timeoutCancel()

		return sttErr
	}, backoffConfig)

	if err != nil {
		err = fmt.Errorf("failed to speech_to_text with Whisper: %v", err)
		logger.Error(err)
		return "", err
	}

	return text, nil
}
