package speech

import (
	"context"
	"fmt"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/sieglu2/virtual-friends-brain/virtualfriends_go"
)

type GoogleTtsClient struct {
	client *texttospeech.Client
}

func NewGoogleTtsClient(ctx context.Context) (*GoogleTtsClient, error) {
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &GoogleTtsClient{
		client: client,
	}, nil
}

func (t *GoogleTtsClient) TextToSpeech(ctx context.Context, text, voiceName string, gender virtualfriends_go.Gender) ([]byte, error) {
	ssmlGender := texttospeechpb.SsmlVoiceGender_NEUTRAL
	switch gender {
	case virtualfriends_go.Gender_Gender_Male:
		ssmlGender = texttospeechpb.SsmlVoiceGender_MALE
	case virtualfriends_go.Gender_Gender_Female:
		ssmlGender = texttospeechpb.SsmlVoiceGender_FEMALE
	default:
		return nil, fmt.Errorf("unsupported gender: %v", gender)
	}

	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			Name:         voiceName,
			LanguageCode: "en-US",
			SsmlGender:   ssmlGender,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_LINEAR16,
			SpeakingRate:  1.1,
		},
	}

	resp, err := t.client.SynthesizeSpeech(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.AudioContent, nil
}
