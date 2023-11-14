package core

import (
	"context"
	"fmt"
	"time"

	"github.com/sieglu2/virtual-friends-brain/common"
	"github.com/sieglu2/virtual-friends-brain/foundation"
	"github.com/sieglu2/virtual-friends-brain/virtualfriends_go"
)

func HandleAccumulateVoiceMessage(ctx context.Context, vfContext *VfContext, request *virtualfriends_go.AccumulateVoiceMessageRequest) {
	logger := foundation.Logger()

	speechToTextStart := time.Now()
	wavBytes := request.VoiceWav
	text, err := SpeechToTextViaWhisper(ctx, vfContext.clients.GetWhisperClient(), wavBytes, 3)
	if err != nil {
		err = fmt.Errorf("failed to process speechToText in HandleAccumulateVoiceMessage: %v", err)
		logger.Error(err)
		vfContext.sendResp(FromError(err))
		return
	}

	speechToTextEnd := time.Now()
	latencyInMs := float64(speechToTextEnd.Sub(speechToTextStart).Milliseconds())
	logger.Infof("speech_to_text.accu latency: %f ms", latencyInMs)
	go func() {
		//if foundation.IsProd() {
		vfContext.clients.GetBigQueryClient().WriteLatencyStats(&common.LatencyStats{
			Env:          foundation.GetEnvironment(),
			SessionId:    vfContext.sessionId,
			UserId:       vfContext.userId,
			UserIp:       vfContext.remoteAddrFromClient,
			CharacterId:  "<accu>",
			LatencyType:  "speech_to_text.accu",
			LatencyValue: latencyInMs,
			Timestamp:    speechToTextEnd,
		})
		//}
	}()

	vfContext.accumulatedMessage.WriteString(text)

	if request.InformReceipt {
		vfResponse := &virtualfriends_go.VfResponse{
			Response: &virtualfriends_go.VfResponse_AccumulateVoiceMessage{
				AccumulateVoiceMessage: &virtualfriends_go.AccumulateVoiceMessageResponse{
					TranscribedText: text,
				},
			},
		}

		_ = vfContext.sendResp(vfResponse)
	}
}
