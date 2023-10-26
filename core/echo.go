package core

import (
	"context"
	"fmt"

	"github.com/sieglu2/virtual-friends-brain/foundation"
	"github.com/sieglu2/virtual-friends-brain/llm"
	"github.com/sieglu2/virtual-friends-brain/virtualfriends_go"
)

func HandleEcho(ctx context.Context, vfContext *VfContext, request *virtualfriends_go.EchoRequest) {
	logger := foundation.Logger()

	reply, action, sentiment := llm.ExtractActionAndSentiment(request.Text)
	replyWav, err := GenerateVoice(ctx, vfContext, reply, request.VoiceConfig)
	if err != nil {
		err = fmt.Errorf("failed to generate voice: %v", err)
		logger.Error(err)
		// let it go through even if we failed to generate the voice.
	}

	response := &virtualfriends_go.EchoResponse{
		Text:      reply,
		Action:    action,
		Sentiment: sentiment,
		ReplyWav:  replyWav,
	}

	vfResponse := &virtualfriends_go.VfResponse{
		Response: &virtualfriends_go.VfResponse_Echo{
			Echo: response,
		},
	}

	_ = vfContext.sendResp(vfResponse)
}
