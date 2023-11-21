package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sieglu2/virtual-friends-brain/common"
	"github.com/sieglu2/virtual-friends-brain/foundation"
	"github.com/sieglu2/virtual-friends-brain/speech"
	"github.com/sieglu2/virtual-friends-brain/virtualfriends_go"
)

const (
	bucketChatHistory = "vf-chat-histories"
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
	case virtualfriends_go.VoiceType_VoiceType_NormalFemale1:
		fallthrough
	case virtualfriends_go.VoiceType_VoiceType_NormalFemale2:
		voiceBytes, err = speech.TextToSpeechWithFlask(ctx, text, virtualfriends_go.Gender_Gender_Female)

	case virtualfriends_go.VoiceType_VoiceType_NormalMale:
		fallthrough
	case virtualfriends_go.VoiceType_VoiceType_Orc:
		voiceBytes, err = speech.TextToSpeechWithFlask(ctx, text, virtualfriends_go.Gender_Gender_Male)

	default:
		return nil, fmt.Errorf("invalid voice_type: %v", voiceConfig.VoiceType)
	}
	if err != nil {
		err = fmt.Errorf("failed to TextToSpeechWithFlask: %v", err)
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

func logChatHistoryToBq(vfContext *VfContext, chatHistory *common.ChatHistory) error {
	logger := foundation.Logger()

	logger.Infof("logging chat history to BQ: %+v", chatHistory)

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

func logChatHistoryToGcs(vfContext *VfContext, chatHistory *common.ChatHistory) error {
	logger := foundation.Logger()

	if len(chatHistory.UserId) == 0 {
		// empty userId, meaning it's an anonymous chat, let's not log the history.
		logger.Infof("empty userId, not gonna log chat history")
		return nil
	}

	logger.Infof("logging chat history to GCS: %+v", chatHistory)

	gcsClient := vfContext.clients.GetGcsClient()
	path := fmt.Sprintf("%s/%s", chatHistory.CharacterId, chatHistory.UserId)

	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	err := gcsClient.SaveContentToGcs(timeoutCtx, bucketChatHistory, path, chatHistory.ChatHistory)
	timeoutCancel()

	if err != nil {
		err = fmt.Errorf("failed to SaveContentToGcs: %v", err)
		logger.Error(err)
		return err
	}

	return nil
}

func loadChatHistoryFromGcs(ctx context.Context, vfContext *VfContext, characterId, userId string) ([]string, error) {
	logger := foundation.Logger()

	if len(userId) == 0 {
		logger.Infof("empty userId, no history is available")
		return []string{}, nil
	}

	path := fmt.Sprintf("%s/%s", characterId, userId)
	rawChatHistory, err := vfContext.clients.GetGcsClient().LoadContentFromGcs(ctx, bucketChatHistory, path)
	if err != nil {
		err = fmt.Errorf("failed to LoadContentFromGcs: %v", err)
		logger.Error(err)
		return []string{}, err
	}

	rawChatHistory = strings.ReplaceAll(rawChatHistory, "\"", "'")
	logger.Infof("rawChatHistory: %s", rawChatHistory)

	splitedChatHistory := strings.Split(rawChatHistory, "\n")
	result := make([]string, 0, len(splitedChatHistory))

	lastRoleName := ""
	messageBuffer := strings.Builder{}

	for lineIndex := 0; lineIndex < len(splitedChatHistory); lineIndex += 1 {
		roleName := ""
		charIndex := 0
		for ; charIndex < len(splitedChatHistory[lineIndex]); charIndex += 1 {
			if splitedChatHistory[lineIndex][charIndex] == ':' {
				roleName = splitedChatHistory[lineIndex][:charIndex]
				break
			}
		}

		if len(roleName) == 0 {
			messageBuffer.WriteString(splitedChatHistory[lineIndex])
		} else {
			if lastRoleName == "A" {
				lastRoleName = "assistant"
			}
			result = append(result, fmt.Sprintf(`{"role":"%s","content":"%s"}`, lastRoleName, messageBuffer.String()))

			lastRoleName = roleName
			messageBuffer.Reset()
			messageBuffer.WriteString(splitedChatHistory[lineIndex][charIndex+1:])
		}
	}

	if lastRoleName == "A" {
		lastRoleName = "assistant"
	}
	result = append(result, fmt.Sprintf(`{"role":"%s","content":"%s"}`, lastRoleName, messageBuffer.String()))

	for _, res := range result {
		logger.Infof("res: %s", res)
	}

	return result, nil
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

func SpeechToTextViaWhisper(ctx context.Context, client *speech.WhisperClient, wavBytes []byte, maxRetries int) (string, error) {
	logger := foundation.Logger()

	var text string
	// TODO: Consider run these DoTries in parallel.
	err := foundation.DoRetry(ctx, func(timeoutCtx context.Context) error {
		var sttErr error
		text, sttErr = client.SpeechToText(timeoutCtx, wavBytes)
		return sttErr
	}, maxRetries, 3000*time.Millisecond)

	if err != nil {
		err = fmt.Errorf("failed to speech_to_text with Whisper: %v", err)
		logger.Error(err)
		return "", err
	}

	return text, nil
}
