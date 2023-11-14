package core

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/sieglu2/virtual-friends-brain/common"
	"github.com/sieglu2/virtual-friends-brain/foundation"
	"github.com/sieglu2/virtual-friends-brain/llm"
	"github.com/sieglu2/virtual-friends-brain/virtualfriends_go"
)

var (
	rpmRegexPattern     = `(?m)^https:\/\/models\.readyplayer\.me\/[0-9a-z]+\.glb$`
	blobDownloadPrefix  = "vf://blob/"
	avaturnRegexPattern = `(?m)^https:\/\/api\.avaturn\.me\/avatars\/exports\/[a-z0-9\-]+\/model$`

	specialCharacters = map[string]string{
		"2bc098d7b8f35d45f86a2f778f5dd89d": "mina",
		"e75d8532c413d425307ef7d42b5ccd94": "einstein",
	}
)

func determineLoader(characterUrl string, response *virtualfriends_go.GetCharacterResponse) error {
	if matched, err := regexp.MatchString(rpmRegexPattern, characterUrl); err == nil && matched {
		response.LoaderConfig = &virtualfriends_go.GetCharacterResponse_LoaderReadyplayerme{
			LoaderReadyplayerme: &virtualfriends_go.LoaderReadyPlayerMe{
				AvatarUrl: characterUrl,
			},
		}
		return nil
	} else if strings.HasPrefix(characterUrl, blobDownloadPrefix) {
		// TODO(yufan.lu), make this changeable too.
		blobName := characterUrl[len(blobDownloadPrefix):]
		response.LoaderConfig = &virtualfriends_go.GetCharacterResponse_LoaderBlobDownload{
			LoaderBlobDownload: &virtualfriends_go.LoaderBlobDownload{
				BlobName:           blobName,
				InBundleObjectName: blobName,
			},
		}
		return nil
	} else if matched, err := regexp.MatchString(avaturnRegexPattern, characterUrl); err == nil && matched {
		response.LoaderConfig = &virtualfriends_go.GetCharacterResponse_LoaderAvaturn{
			LoaderAvaturn: &virtualfriends_go.LoaderAvaturn{
				AvatarUrl: characterUrl,
			},
		}
		return nil
	}

	return fmt.Errorf("unsupported character url: %s", characterUrl)
}

func HandleGetCharacter(ctx context.Context, vfContext *VfContext, request *virtualfriends_go.GetCharacterRequest) {
	logger := foundation.Logger()

	// vfResponse := &virtualfriends_go.VfResponse{}
	response := &virtualfriends_go.GetCharacterResponse{}

	// Generate a new UUID
	newUUID, err := uuid.NewRandom()
	if err != nil {
		err = fmt.Errorf("failed to generate uuid: %v", err)
		logger.Error(err)
		vfContext.sendResp(FromError(err))
		return
	}

	response.GeneratedSessionId = newUUID.String()

	lowerCharacterId := strings.ToLower(request.CharacterId)
	if lowerCharacterId == "mina" {
		if err := determineLoader("vf://blob/mina", response); err != nil {
			err = fmt.Errorf("failed to determine loader for mina: %v", err)
			logger.Error(err)
			vfContext.sendResp(FromError(err))
			return
		}

		response.VoiceConfig = &virtualfriends_go.VoiceConfig{
			VoiceType: virtualfriends_go.VoiceType_VoiceType_NormalFemale2,
		}
		response.Gender = virtualfriends_go.Gender_Gender_Female
		response.FriendName = "mina"

		response.Greeting = "Hi there, I'm Mina, an AI assistant created by the Virtual Friends team. What can I help you?"
		voiceBytes, err := GenerateVoice(ctx, vfContext, response.Greeting, response.VoiceConfig)
		if err != nil {
			err = fmt.Errorf("failed to GenerateVoice for mina: %v", err)
			logger.Error(err)
			vfContext.sendResp(FromError(err))
			return
		}
		response.GreetingWav = voiceBytes
		response.Description = "Mina is an AI assistant created by the Virtual Friends team."
		response.BasePrompts = strings.Join([]string{
			"Mina is an AI assistant created by the Virtual Friends team.\n",
			"Here is information about the Virtual Friends project:\n",
			common.VirtualFriendsInfo,
			"\n---\n",
			"Act as Mina.\n",
			"Be concise in your response; do not provide extensive information at once.",
		}, "")
	} else if lowerCharacterId == "einstein" {
		if err := determineLoader("vf://blob/einstein", response); err != nil {
			err = fmt.Errorf("failed to determine loader for einstein: %v", err)
			logger.Error(err)
			vfContext.sendResp(FromError(err))
			return
		}

		response.VoiceConfig = &virtualfriends_go.VoiceConfig{
			VoiceType: virtualfriends_go.VoiceType_VoiceType_NormalMale,
		}
		response.Gender = virtualfriends_go.Gender_Gender_Male
		response.FriendName = "einstein"

		response.Greeting = "Hello, I'm Einstein, a passionate scientist by day and an ardent stargazer by night."
		voiceBytes, err := GenerateVoice(ctx, vfContext, response.Greeting, response.VoiceConfig)
		if err != nil {
			err = fmt.Errorf("failed to GenerateVoice for einstein: %v", err)
			logger.Error(err)
			vfContext.sendResp(FromError(err))
			return
		}
		response.GreetingWav = voiceBytes
		response.Description = "Einstein is one of the most famous scientists who changed the whole world."
		response.BasePrompts = common.ExampleCharacterPrompts["einstein"]
	} else {
		character, err := vfContext.clients.GetDatastoreClient().QueryCharacter(ctx, request.CharacterId)

		vfContext.assistantId = character.AssistantId
		vfContext.openaiApiKey = character.OpenaiApiKey

		// TODO(ysong): support user's openai api key
		threadId, err := llm.CreateThreadWithFlask(ctx)
		if err != nil {
			err = fmt.Errorf("Error creating openai assistant thread:: %v", err)
			logger.Error(err)
			vfContext.sendResp(FromError(err))
			return
		}
		vfContext.threadId = threadId

		if err != nil {
			err = fmt.Errorf("failed to QueryCharacter: %v", err)
			logger.Error(err)
			vfContext.sendResp(FromError(err))
			return
		}

		err = vfContext.clients.GetGcsClient().ExtendCharacterInfo(ctx, character)
		if err != nil {
			err = fmt.Errorf("failed to ExtendCharacterInfo: %v", err)
			logger.Error(err)
			vfContext.sendResp(FromError(err))
			return
		}

		if err := determineLoader(character.AvatarUrl, response); err != nil {
			err = fmt.Errorf("failed to determine loader for %s: %v", request.CharacterId, err)
			logger.Error(err)
			vfContext.sendResp(FromError(err))
			return
		}

		voiceConfig := &virtualfriends_go.VoiceConfig{}
		switch strings.ToLower(character.Gender) {
		case "male":
			response.Gender = virtualfriends_go.Gender_Gender_Male
			voiceConfig.VoiceType = virtualfriends_go.VoiceType_VoiceType_NormalMale
		case "female":
			response.Gender = virtualfriends_go.Gender_Gender_Female
			voiceConfig.VoiceType = virtualfriends_go.VoiceType_VoiceType_NormalFemale2
		}
		if val, exist := specialCharacters[request.CharacterId]; exist {
			switch val {
			case "mina":
				voiceConfig.Octaves = 0.18

			case "einstein":
				voiceConfig.Octaves = -0.2
			}
		}
		voiceConfig.ElevenLabId = character.ElevenLabId
		response.VoiceConfig = voiceConfig

		response.FriendName = character.Name
		if len(response.FriendName) == 0 {
			response.FriendName = "Virtual Friends Assistant"
		}
		response.Greeting = strings.ReplaceAll(character.Greeting, "\"", "\\\"")
		if len(response.Greeting) == 0 {
			response.Greeting = "hi, I am Virtual Friends Assistant."
		}

		response.Description = strings.ReplaceAll(character.Description, "\"", "\\\"")
		response.BasePrompts = strings.Join([]string{
			fmt.Sprintf("name: %s", response.FriendName),
			fmt.Sprintf("description: %s", response.Description),
			strings.ReplaceAll(character.Prompts, "\"", "\\\""),
			fmt.Sprintf("Act as %s", response.FriendName),
		}, "\n")

		voiceBytes, err := GenerateVoice(ctx, vfContext, response.Greeting, response.VoiceConfig)
		if err != nil {
			err = fmt.Errorf("failed to GenerateVoice: %v", err)
			logger.Error(err)
			vfContext.sendResp(FromError(err))
			return
		}
		response.GreetingWav = voiceBytes
	}

	vfResponse := &virtualfriends_go.VfResponse{
		Response: &virtualfriends_go.VfResponse_GetCharacter{
			GetCharacter: response,
		},
	}
	_ = vfContext.sendResp(vfResponse)
}
