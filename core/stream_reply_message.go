package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sieglu2/virtual-friends-brain/common"
	"github.com/sieglu2/virtual-friends-brain/foundation"
	"github.com/sieglu2/virtual-friends-brain/llm"
	"github.com/sieglu2/virtual-friends-brain/virtualfriends_go"
)

const (
	splitChars       = ".;!?:,。；！？：，"
	replyTextOnError = "sorry I'm having some troubles, can you try say it again? [no_action] <neutral>"

	noMedicalQuestions = "Do not answer any medical related questions.\n"
)

func HandleStreamReplyMessage(ctx context.Context, vfContext *VfContext, request *virtualfriends_go.StreamReplyMessageRequest) {
	logger := foundation.Logger()

	text := ""
	var err error
	switch request.CurrentMessage.(type) {
	case *virtualfriends_go.StreamReplyMessageRequest_UseAccumulated:
		text = vfContext.accumulatedMessage.String()
		vfContext.accumulatedMessage.Reset()

	case *virtualfriends_go.StreamReplyMessageRequest_Wav:
		speechToTextStart := time.Now()
		wavBytes := request.CurrentMessage.(*virtualfriends_go.StreamReplyMessageRequest_Wav).Wav
		text, err = SpeechToTextViaWhisper(ctx, vfContext.clients.GetWhisperClient(), wavBytes, 3)
		if err != nil {
			err = fmt.Errorf("failed to process speechToText in HandleStreamReplyMessage: %v", err)
			logger.Error(err)
			sendReply(ctx, vfContext, request, text, replyTextOnError, 0, false)
			sendStopReply(ctx, vfContext, request, 1)
			return
		}

		speechToTextEnd := time.Now()
		latencyInMs := float64(speechToTextEnd.Sub(speechToTextStart).Milliseconds())
		logger.Infof("speech_to_text.stream latency: %f ms", latencyInMs)
		go func() {
			if foundation.IsProd() {
				vfContext.clients.GetBigQueryClient().WriteLatencyStats(&common.LatencyStats{
					Env:          foundation.GetEnvironment(),
					SessionId:    vfContext.sessionId,
					UserId:       vfContext.userId,
					UserIp:       vfContext.remoteAddrFromClient,
					CharacterId:  request.MirroredContent.CharacterId,
					LatencyType:  "speech_to_text.stream",
					LatencyValue: latencyInMs,
					Timestamp:    speechToTextEnd,
				})
			}
		}()
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

	basePrompts = noMedicalQuestions + basePrompts
	basePrompts = strings.ReplaceAll(basePrompts, "\r", "")
	basePrompts = strings.ReplaceAll(basePrompts, "\n", "\\n")

	firstJson := fmt.Sprintf(`{"role":"system","content":"%s"}`, basePrompts)
	lastJson := fmt.Sprintf(`{"role":"user","content":"%s"}`, currentMessage)
	chronicalJsons = append([]string{firstJson}, chronicalJsons...)
	chronicalJsons = append(chronicalJsons, lastJson)

	vfContext.savedCharacterId = request.MirroredContent.CharacterId
	vfContext.savedJsonMessages = chronicalJsons[1:]

	buffer := strings.Builder{}
	replyIndex := 0

	completeReply := strings.Builder{}
	llmInferStart := time.Now()
	processStreamText := func(replyText string, index int) error {
		if len(replyText) == 0 {
			return nil
		}

		if index == 0 {
			llmInferEnd := time.Now()
			latencyInMs := float64(llmInferEnd.Sub(llmInferStart).Milliseconds())
			logger.Infof("llm_infer latency: %f ms", latencyInMs)
			go func() {
				if foundation.IsProd() {
					vfContext.clients.GetBigQueryClient().WriteLatencyStats(&common.LatencyStats{
						Env:          foundation.GetEnvironment(),
						SessionId:    vfContext.sessionId,
						UserId:       vfContext.userId,
						UserIp:       vfContext.remoteAddrFromClient,
						CharacterId:  request.MirroredContent.CharacterId,
						LatencyType:  "llm_infer",
						LatencyValue: latencyInMs,
						Timestamp:    llmInferEnd,
					})
				}
			}()
		}

		buffer.WriteString(replyText)
		completeReply.WriteString(replyText)
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

	if completeReply.Len() > 0 {
		completeReplyStr := completeReply.String()
		reply, _, _ := llm.ExtractActionAndSentiment(completeReplyStr)
		vfContext.savedJsonMessages = append(
			vfContext.savedJsonMessages,
			fmt.Sprintf(`{"role":"assistant","content":"%s"}`, reply))
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

	if !isStop {
		if len(rawReplyText) > 0 {
			reply, action, sentiment := llm.ExtractActionAndSentiment(rawReplyText)
			logger.Infof("reply: %s, action: %s, sentiment: %s", reply, action, sentiment)

			if len(reply) > 0 {
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
			}

			response.Action = action
			response.Sentiment = sentiment
		}
	}

	vfResponse := &virtualfriends_go.VfResponse{
		Response: &virtualfriends_go.VfResponse_StreamReplyMessage{
			StreamReplyMessage: response,
		},
	}

	_ = vfContext.sendResp(vfResponse)

	return nil
}
