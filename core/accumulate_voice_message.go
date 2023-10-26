package core

import (
	"context"
	"fmt"
	"time"

	"github.com/sieglu2/virtual-friends-brain/common"
	"github.com/sieglu2/virtual-friends-brain/foundation"
	"github.com/sieglu2/virtual-friends-brain/speech"
	"github.com/sieglu2/virtual-friends-brain/virtualfriends_go"
)

func HandleAccumulateVoiceMessage(ctx context.Context, vfContext *VfContext, request *virtualfriends_go.AccumulateVoiceMessageRequest) {
	logger := foundation.Logger()

	speechToTextStart := time.Now()
	wavBytes := request.VoiceWav
	text, err := speech.SpeechToText(ctx, wavBytes)
	if err != nil {
		err = fmt.Errorf("failed to process speechToText in HandleAccumulateVoiceMessage: %v", err)
		logger.Error(err)
		vfContext.sendResp(FromError(err))
		return
	}

	go func() {
		if foundation.IsProd() {
			vfContext.clients.GetBigQueryClient().WriteLatencyStats(&common.LatencyStats{
				Env:          foundation.GetEnvironment(),
				SessionId:    vfContext.originalVfRequest.SessionId,
				UserId:       vfContext.originalVfRequest.UserId,
				UserIp:       vfContext.remoteAddr,
				CharacterId:  "<accu>",
				LatencyType:  "speech_to_text.accu",
				LatencyValue: float64(time.Now().Sub(speechToTextStart).Milliseconds()),
				Timestamp:    time.Now(),
			})
		}
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
