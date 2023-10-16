package core

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
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

func HandleGetCharacter(request *virtualfriends_go.GetCharacterRequest, vfContext *VfContext) error {
	logger := foundation.Logger()

	// vfResponse := &virtualfriends_go.VfResponse{}
	response := &virtualfriends_go.GetCharacterResponse{}

	// Generate a new UUID
	newUUID, err := uuid.NewRandom()
	if err != nil {
		fmt.Println("Error generating UUID:", err)
		return err
	}

	response.GeneratedSessionId = newUUID.String()

	if strings.ToLower(request.CharacterId) == "mina" {
		if err := determineLoader("vf://blob/mina", response); err != nil {
			err = fmt.Errorf("failed to determin loader for mina: %v", err)
			logger.Error(err)
			return err
		}

		response.VoiceConfig = &virtualfriends_go.VoiceConfig{
			VoiceType: virtualfriends_go.VoiceType_VoiceType_NormalFemale2,
		}
		response.Gender = virtualfriends_go.Gender_Gender_Female
		response.FriendName = "mina"

		response.Greeting = "Hi there, I'm Mina, an AI assistant created by the Virtual Friends team. What can I help you?"

	}
	return nil
}
