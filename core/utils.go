package core

import (
	"context"
	"fmt"

	"github.com/sieglu2/virtual-friends-brain/speech"
	"github.com/sieglu2/virtual-friends-brain/virtualfriends_go"
)

func GenerateVoice(ctx context.Context, vfContext *VfContext, text string, voiceConfig *virtualfriends_go.VoiceConfig) ([]byte, error) {
	if len(voiceConfig.ElevenLabId) > 0 {
		// TODO (yufan.lu), plug in 11Lab API
		return []byte{}, nil
	}

	var voiceBytes []byte
	var err error
	switch voiceConfig.VoiceType {
	case virtualfriends_go.VoiceType_VoiceType_NormalMale:
		voiceBytes, err = vfContext.clients.GetGoogleTtsClient().TextToSpeech(ctx, text, "en-US-News-M", virtualfriends_go.Gender_Gender_Male)
	case virtualfriends_go.VoiceType_VoiceType_NormalFemale1:
		voiceBytes, err = vfContext.clients.GetGoogleTtsClient().TextToSpeech(ctx, text, "en-US-News-K", virtualfriends_go.Gender_Gender_Female)
	case virtualfriends_go.VoiceType_VoiceType_NormalFemale2:
		voiceBytes, err = vfContext.clients.GetGoogleTtsClient().TextToSpeech(ctx, text, "en-US-News-L", virtualfriends_go.Gender_Gender_Female)
	case virtualfriends_go.VoiceType_VoiceType_Orc:
		voiceBytes, err = vfContext.clients.GetGoogleTtsClient().TextToSpeech(ctx, text, "en-US-News-M", virtualfriends_go.Gender_Gender_Male)
	default:
		return nil, fmt.Errorf("invalid voice_type: %v", voiceConfig.VoiceType)
	}
	if err != nil {
		return nil, err
	}

	if voiceConfig.Octaves == 0 {
		return voiceBytes, nil
	}
	return speech.TransposeSound(voiceBytes, float64(voiceConfig.Octaves))
}
