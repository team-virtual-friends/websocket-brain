package speech

import (
	"context"
	"fmt"
	"strings"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/sieglu2/virtual-friends-brain/virtualfriends_go"
)

type GoogleTtsConfig struct {
	fullLanguageCode                            string
	femaleVoiceName, alternativeFemaleVoiceName string
	maleVoiceName                               string
}

var (
	UnsupportedLanguagePrefix = "unsupported language:"

	languageCodeMap = map[string]GoogleTtsConfig{
		"ar": GoogleTtsConfig{
			fullLanguageCode:           "ar-XA",
			femaleVoiceName:            "ar-XA-Standard-A",
			alternativeFemaleVoiceName: "ar-XA-Standard-D",
			maleVoiceName:              "ar-XA-Standard-B",
		},
		"bn": GoogleTtsConfig{
			fullLanguageCode:           "bn-IN",
			femaleVoiceName:            "bn-IN-Standard-A",
			alternativeFemaleVoiceName: "bn-IN-Wavenet-A",
			maleVoiceName:              "bn-IN-Standard-B",
		},
		"de": GoogleTtsConfig{
			fullLanguageCode:           "de-DE",
			femaleVoiceName:            "de-DE-Standard-A",
			alternativeFemaleVoiceName: "de-DE-Standard-C",
			maleVoiceName:              "de-DE-Standard-B",
		},
		"el": GoogleTtsConfig{
			fullLanguageCode:           "el-GR",
			femaleVoiceName:            "el-GR-Standard-A",
			alternativeFemaleVoiceName: "el-GR-Wavenet-A",
			maleVoiceName:              "",
		},
		"en": GoogleTtsConfig{
			fullLanguageCode:           "en-US",
			femaleVoiceName:            "en-US-Standard-H",
			alternativeFemaleVoiceName: "en-US-Neural2-F",
			maleVoiceName:              "en-US-Standard-B",
		},
		"es": GoogleTtsConfig{
			fullLanguageCode:           "es-ES",
			femaleVoiceName:            "es-ES-Standard-A",
			alternativeFemaleVoiceName: "es-ES-Standard-C",
			maleVoiceName:              "es-ES-Standard-B",
		},
		"fr": GoogleTtsConfig{
			fullLanguageCode:           "fr-FR",
			femaleVoiceName:            "fr-FR-Standard-A",
			alternativeFemaleVoiceName: "fr-FR-Standard-C",
			maleVoiceName:              "fr-FR-Standard-B",
		},
		"he": GoogleTtsConfig{
			fullLanguageCode:           "he-IL",
			femaleVoiceName:            "he-IL-Standard-A",
			alternativeFemaleVoiceName: "he-IL-Standard-C",
			maleVoiceName:              "he-IL-Standard-B",
		},
		"hi": GoogleTtsConfig{
			fullLanguageCode:           "hi-IN",
			femaleVoiceName:            "hi-IN-Standard-A",
			alternativeFemaleVoiceName: "hi-IN-Standard-D",
			maleVoiceName:              "hi-IN-Standard-B",
		},
		"hu": GoogleTtsConfig{
			fullLanguageCode:           "hu-HU",
			femaleVoiceName:            "hu-HU-Standard-A",
			alternativeFemaleVoiceName: "hu-HU-Wavenet-A",
			maleVoiceName:              "",
		},
		// "hy"
		"gu": GoogleTtsConfig{
			fullLanguageCode:           "gu-IN",
			femaleVoiceName:            "gu-IN-Standard-A",
			alternativeFemaleVoiceName: "gu-IN-Wavenet-A",
			maleVoiceName:              "gu-IN-Standard-B",
		},
		"it": GoogleTtsConfig{
			fullLanguageCode:           "it-IT",
			femaleVoiceName:            "it-IT-Standard-A",
			alternativeFemaleVoiceName: "it-IT-Standard-B",
			maleVoiceName:              "it-IT-Standard-C",
		},
		"nl": GoogleTtsConfig{
			fullLanguageCode:           "nl-NL",
			femaleVoiceName:            "nl-NL-Standard-A",
			alternativeFemaleVoiceName: "nl-NL-Standard-D",
			maleVoiceName:              "nl-NL-Standard-B",
		},
		"pl": GoogleTtsConfig{
			fullLanguageCode:           "pl-PL",
			femaleVoiceName:            "pl-PL-Standard-A",
			alternativeFemaleVoiceName: "pl-PL-Standard-D",
			maleVoiceName:              "pl-PL-Standard-B",
		},
		"pt": GoogleTtsConfig{
			fullLanguageCode:           "pt-BR",
			femaleVoiceName:            "pt-BR-Standard-A",
			alternativeFemaleVoiceName: "pt-BR-Wavenet-A",
			maleVoiceName:              "pt-BR-Standard-B",
		},
		"pa": GoogleTtsConfig{
			fullLanguageCode:           "pa-IN",
			femaleVoiceName:            "pa-IN-Standard-A",
			alternativeFemaleVoiceName: "pa-IN-Standard-C",
			maleVoiceName:              "pa-IN-Standard-B",
		},
		"ja": GoogleTtsConfig{
			fullLanguageCode:           "ja-JP",
			femaleVoiceName:            "ja-JP-Standard-A",
			alternativeFemaleVoiceName: "ja-JP-Standard-B",
			maleVoiceName:              "ja-JP-Standard-C",
		},
		// "ka"
		"ko": GoogleTtsConfig{
			fullLanguageCode:           "ko-KR",
			femaleVoiceName:            "ko-KR-Standard-A",
			alternativeFemaleVoiceName: "ko-KR-Standard-B",
			maleVoiceName:              "ko-KR-Standard-C",
		},
		"ta": GoogleTtsConfig{
			fullLanguageCode:           "ta-IN",
			femaleVoiceName:            "ta-IN-Standard-A",
			alternativeFemaleVoiceName: "ta-IN-Standard-C",
			maleVoiceName:              "ta-IN-Standard-B",
		},
		"te": GoogleTtsConfig{
			fullLanguageCode:           "te-IN",
			femaleVoiceName:            "te-IN-Standard-A",
			alternativeFemaleVoiceName: "te-IN-Standard-A",
			maleVoiceName:              "te-IN-Standard-B",
		},
		// "tl"
		"th": GoogleTtsConfig{
			fullLanguageCode:           "th-TH",
			femaleVoiceName:            "th-TH-Standard-A",
			alternativeFemaleVoiceName: "th-TH-Neural2-C",
			maleVoiceName:              "",
		},
		"ru": GoogleTtsConfig{
			fullLanguageCode:           "ru-RU",
			femaleVoiceName:            "ru-RU-Standard-A",
			alternativeFemaleVoiceName: "ru-RU-Standard-C",
			maleVoiceName:              "ru-RU-Standard-B",
		},
		"sr": GoogleTtsConfig{
			fullLanguageCode:           "sr-RS",
			femaleVoiceName:            "sr-RS-Standard-A",
			alternativeFemaleVoiceName: "sr-RS-Standard-A",
			maleVoiceName:              "",
		},
		"vi": GoogleTtsConfig{
			fullLanguageCode:           "vi-VN",
			femaleVoiceName:            "vi-VN-Standard-A",
			alternativeFemaleVoiceName: "vi-VN-Standard-C",
			maleVoiceName:              "vi-VN-Standard-B",
		},
		"uk": GoogleTtsConfig{
			fullLanguageCode:           "uk-UA",
			femaleVoiceName:            "uk-UA-Standard-A",
			alternativeFemaleVoiceName: "uk-UA-Wavenet-A",
			maleVoiceName:              "",
		},
		"zh": GoogleTtsConfig{
			fullLanguageCode:           "cmn-CN",
			femaleVoiceName:            "cmn-CN-Standard-A",
			alternativeFemaleVoiceName: "cmn-CN-Standard-D",
			maleVoiceName:              "cmn-CN-Standard-C",
		},
	}
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

func (t *GoogleTtsClient) TextToSpeech(ctx context.Context, text string, gender virtualfriends_go.Gender, withAlternativeFemaleVoice bool) ([]byte, error) {
	shortLangCode := DetectShortLanguageCode(text)

	fullLanguageCode, err := getFullLanguageCode(shortLangCode)
	if err != nil {
		return nil, err
	}

	var voiceName string
	ssmlGender := texttospeechpb.SsmlVoiceGender_NEUTRAL
	switch gender {
	case virtualfriends_go.Gender_Gender_Male:
		ssmlGender = texttospeechpb.SsmlVoiceGender_MALE
		voiceName, err = getMaleVoiceName(shortLangCode)

	case virtualfriends_go.Gender_Gender_Female:
		ssmlGender = texttospeechpb.SsmlVoiceGender_FEMALE
		voiceName, err = getFemaleVoiceName(shortLangCode, withAlternativeFemaleVoice)

	default:
		return nil, fmt.Errorf("unsupported gender: %v", gender)
	}

	if err != nil {
		return nil, err
	}

	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			Name:         voiceName,
			LanguageCode: fullLanguageCode,
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

func getFullLanguageCode(shortLangCode string) (string, error) {
	if val, exist := languageCodeMap[shortLangCode]; exist {
		return val.fullLanguageCode, nil
	}
	return "", unsupportedLangaugeError(shortLangCode)
}

func getFemaleVoiceName(shortLangCode string, isAlternative bool) (string, error) {
	if val, exist := languageCodeMap[shortLangCode]; exist {
		voiceName := val.femaleVoiceName
		if isAlternative {
			voiceName = val.alternativeFemaleVoiceName
		}
		if len(voiceName) == 0 {
			return "", unsupportedLangaugeError(shortLangCode)
		}
		return voiceName, nil
	}
	return "", unsupportedLangaugeError(shortLangCode)
}

func getMaleVoiceName(shortLangCode string) (string, error) {
	if val, exist := languageCodeMap[shortLangCode]; exist {
		voiceName := val.maleVoiceName
		if len(voiceName) == 0 {
			return "", unsupportedLangaugeError(shortLangCode)
		}
		return voiceName, nil
	}
	return "", unsupportedLangaugeError(shortLangCode)
}

func unsupportedLangaugeError(shortLangCode string) error {
	return fmt.Errorf(UnsupportedLanguagePrefix + shortLangCode)
}

func IsUnsupportedLanguageError(err error) bool {
	return strings.HasPrefix(err.Error(), UnsupportedLanguagePrefix)
}
