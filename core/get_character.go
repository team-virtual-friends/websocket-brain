package core

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/sieglu2/virtual-friends-brain/common"
	"github.com/sieglu2/virtual-friends-brain/foundation"
	"github.com/sieglu2/virtual-friends-brain/virtualfriends_go"
)

var (
	rpmRegexPattern     = `(?m)^https:\/\/models\.readyplayer\.me\/[0-9a-z]+\.glb$`
	blobDownloadPrefix  = "vf://blob/"
	avaturnRegexPattern = `(?m)^https:\/\/api\.avaturn\.me\/avatars\/exports\/[a-z0-9\-]+\/model$`
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
	} else if regexp.MatchString(avaturnRegexPattern, characterUrl); err == nil && matched {
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
			err = fmt.Errorf("failed to GenerateVoice: %v", err)
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
			err = fmt.Errorf("failed to GenerateVoice: %v", err)
			logger.Error(err)
			vfContext.sendResp(FromError(err))
			return
		}
		response.GreetingWav = voiceBytes
		response.Description = "Einstein is one of the most famous scientists who changed the whole world."
		response.BasePrompts = common.ExampleCharacterPrompts["einstein"]
	} else {
		// TODO (yufan.lu) datastore lookup.
	}

	vfResponse := &virtualfriends_go.VfResponse{
		Response: &virtualfriends_go.VfResponse_GetCharacter{
			GetCharacter: response,
		},
	}
	_ = vfContext.sendResp(vfResponse)
}
