package core

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/sieglu2/virtual-friends-brain/foundation"
	"github.com/sieglu2/virtual-friends-brain/llm"
	"github.com/sieglu2/virtual-friends-brain/virtualfriends_go"
)

const (
	splitChars       = ".;!?:,。；！？：，"
	replyTextOnError = "sorry I'm having some troubles, can you try say it again? [no_action] <neutral>"
)

func HandleStreamReplyMessage(ctx context.Context, vfContext *VfContext, request *virtualfriends_go.StreamReplyMessageRequest) {
	logger := foundation.Logger()

	text := ""
	var err error
	switch request.CurrentMessage.(type) {
	case *virtualfriends_go.StreamReplyMessageRequest_Wav:
		wavBytes := request.CurrentMessage.(*virtualfriends_go.StreamReplyMessageRequest_Wav).Wav
		text, err = speechToText(ctx, wavBytes)
		if err != nil {
			err = fmt.Errorf("failed to process speechToText: %v", err)
			logger.Error(err)
			sendReply(ctx, vfContext, request, text, replyTextOnError, 0, false)
			sendStopReply(ctx, vfContext, request, 1)
			return
		}
		logger.Infof("speechToText result: %s", text)

	case *virtualfriends_go.StreamReplyMessageRequest_Text:
		text = request.CurrentMessage.(*virtualfriends_go.StreamReplyMessageRequest_Text).Text
	}

	err = llmStreamReply(ctx, vfContext, request, request.BasePrompts, text, request.JsonMessages)
	if err != nil {
		err = fmt.Errorf("failed to process speechToText: %v", err)
		logger.Error(err)
		sendReply(ctx, vfContext, request, text, replyTextOnError, 0, false)
		sendStopReply(ctx, vfContext, request, 1)
		return
	}

	logger.Info("done streaming reply")
}

func speechToText(ctx context.Context, wavBytes []byte) (string, error) {
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

func llmStreamReply(
	ctx context.Context, vfContext *VfContext, request *virtualfriends_go.StreamReplyMessageRequest,
	basePrompts string, currentMessage string, chronicalJsons []string,
) error {
	logger := foundation.Logger()
	var err error

	if len(basePrompts) == 0 {
		basePrompts = "You are an AI assistant created by Virtual Friends Team."
	} else {
		basePrompts += llm.InferSentimentAndActionPrompt
	}

	basePrompts = strings.ReplaceAll(basePrompts, "\r", "")
	basePrompts = strings.ReplaceAll(basePrompts, "\n", "\\n")
	firstJson := fmt.Sprintf(`{"role":"system","content":"%s"}`, basePrompts)
	lastJson := fmt.Sprintf(`{"role":"user","content":"%s"}`, currentMessage)
	chronicalJsons = append([]string{firstJson}, chronicalJsons...)
	chronicalJsons = append(chronicalJsons, lastJson)

	buffer := strings.Builder{}
	replyIndex := 0

	processStreamText := func(replyText string) error {
		if len(replyText) == 0 {
			return nil
		}

		buffer.WriteString(replyText)
		if strings.Contains(splitChars, string(replyText[len(replyText)-1])) {
			bufferString := buffer.String()
			err := sendReply(ctx, vfContext, request, currentMessage, bufferString, replyIndex, false)
			if err != nil {
				err = fmt.Errorf("failed to sendReply(%s): %v", bufferString, err)
				logger.Error(err)
				return err
			}
			replyIndex += 1
			buffer.Reset()
		}

		return nil
	}

	completionStreamText := func() error {
		bufferString := buffer.String()
		if len(bufferString) > 0 {
			err := sendReply(ctx, vfContext, request, currentMessage, bufferString, replyIndex, false)
			if err != nil {
				err = fmt.Errorf("failed to sendReply(%s): %v", bufferString, err)
				logger.Error(err)
				return err
			}
		}
		sendStopReply(ctx, vfContext, request, replyIndex+1)

		return nil
	}

	err = vfContext.clients.GetChatGptClient().StreamReplyMessage(ctx, chronicalJsons, processStreamText, completionStreamText)
	if err != nil {
		err = fmt.Errorf("failed to StreamReplyMessage from llm: %v", err)
		logger.Error(err)
		return err
	}

	return nil
}

func sendStopReply(
	ctx context.Context, vfContext *VfContext,
	request *virtualfriends_go.StreamReplyMessageRequest, replyIndex int,
) {
	sendReply(ctx, vfContext, request, "", "", replyIndex, true)
}

func sendReply(
	ctx context.Context,
	vfContext *VfContext, request *virtualfriends_go.StreamReplyMessageRequest,
	currentMessage string, rawReplyText string, replyIndex int, isStop bool,
) error {
	logger := foundation.Logger()
	rawReplyText = strings.Trim(rawReplyText, " ")

	response := &virtualfriends_go.StreamReplyMessageResponse{
		MirroredContent: request.MirroredContent,
		ChunkIndex:      int32(replyIndex),
		IsStop:          isStop,
	}

	if len(rawReplyText) > 0 {
		reply, action, sentiment := llm.ExtractActionAndSentiment(rawReplyText)
		logger.Infof("reply: %s, action: %s, sentiment: %s", reply, action, sentiment)

		response.ReplyMessage = reply
		if replyIndex == 0 {
			response.TranscribedText = currentMessage
		}

		replyWav, err := GenerateVoice(ctx, vfContext, reply, request.VoiceConfig)
		if err != nil {
			err = fmt.Errorf("failed to generate voice: %v", err)
			logger.Error(err)
			// let it go through even if we failed to generate the voice.
		}
		response.ReplyWav = replyWav

		response.Action = action
		response.Sentiment = sentiment
	}

	vfResponse := virtualfriends_go.VfResponse{
		Response: &virtualfriends_go.VfResponse_StreamReplyMessage{
			StreamReplyMessage: response,
		},
	}

	_ = vfContext.sendResp(&vfResponse)

	return nil
}
