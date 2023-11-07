package core

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sieglu2/virtual-friends-brain/common"
	"github.com/sieglu2/virtual-friends-brain/foundation"
	"github.com/sieglu2/virtual-friends-brain/virtualfriends_go"
	"google.golang.org/protobuf/proto"
)

const (
	VirtualFriendsOrigin = "https://virtualfriends.ai"
	prefixLocalHost      = "http://localhost:"
	prefix127            = "http://127.0.0.1:"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		logger := foundation.Logger()

		origins := r.Header["Origin"]
		if len(origins) == 0 {
			logger.Errorf("no origin given")
			return false
		}
		logger.Infof("origins from header: %v", origins)
		for _, origin := range origins {
			trimmed := strings.Trim(origin, " ")
			if trimmed == VirtualFriendsOrigin ||
				strings.HasPrefix(trimmed, prefixLocalHost) ||
				strings.HasPrefix(trimmed, prefix127) {

				return true
			}
		}
		return false
	},
} // use default options

type VfContext struct {
	// use this delegate to send VfResponse to client single or multiple times.
	sendResp func(vfResponse *virtualfriends_go.VfResponse) error

	// those are updated from VfRequest.
	remoteAddrFromClient string
	userId               string
	sessionId            string
	runtimeEnv           virtualfriends_go.RuntimeEnv

	// these are updated via HandleStreamReplyMessage.
	savedCharacterId  string
	savedJsonMessages []string

	clients *common.Clients

	// for saving the accumulated message transcribed in HandleAccumulateVoiceMessage
	// and to be used in StreamReply.
	accumulatedMessage strings.Builder
}

func FromError(err error) *virtualfriends_go.VfResponse {
	return &virtualfriends_go.VfResponse{
		Error: &virtualfriends_go.CustomError{
			ErrorMessage: err.Error(),
		},
	}
}

func OnConnect(conn *websocket.Conn) *VfContext {
	logger := foundation.Logger()

	return &VfContext{
		sendResp: func(vfResponse *virtualfriends_go.VfResponse) error {
			vfResponseBytes, err := proto.Marshal(vfResponse)
			if err != nil {
				logger.Errorf("failed to marshal: %v", err)
				return err
			}
			return conn.WriteMessage(websocket.BinaryMessage, vfResponseBytes)
		},

		clients: common.GetGlobalClients(),
	}
}

func OnDisconnect(vfContext *VfContext) {
	logger := foundation.Logger()

	if foundation.IsProd() && len(vfContext.savedJsonMessages) > 0 {
		disconnectTime := time.Now()
		go logChatHistory(vfContext, &common.ChatHistory{
			UserId:        vfContext.userId,
			UserIp:        vfContext.remoteAddrFromClient,
			CharacterId:   vfContext.savedCharacterId,
			ChatHistory:   assembleChatHistory(vfContext.savedJsonMessages),
			Timestamp:     disconnectTime,
			ChatSessionId: vfContext.sessionId,
			RuntimeEnv:    vfContext.runtimeEnv.String(),
		}, disconnectTime)
	}
	logger.Infof("disconnected.")
}

func InGame(w http.ResponseWriter, r *http.Request) {
	logger := foundation.Logger()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("failed to upgrade: %v", err)
		return
	}
	defer conn.Close()

	vfContext := OnConnect(conn)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			OnDisconnect(vfContext)
			break
		}

		vfRequest := &virtualfriends_go.VfRequest{}
		if err = proto.Unmarshal(message, vfRequest); err != nil {
			logger.Errorf("failed to unmarshal: %v", err)
			break
		}

		vfContext.userId = vfRequest.UserId
		vfContext.sessionId = vfRequest.SessionId
		vfContext.runtimeEnv = vfRequest.RuntimeEnv
		vfContext.remoteAddrFromClient = vfRequest.IpAddr

		handlingCtx, handlingCancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer handlingCancel()

		switch vfRequest.Request.(type) {
		case *virtualfriends_go.VfRequest_Echo:
			request := vfRequest.Request.(*virtualfriends_go.VfRequest_Echo).Echo
			HandleEcho(handlingCtx, vfContext, request)
			return

		case *virtualfriends_go.VfRequest_StreamReplyMessage:
			request := vfRequest.Request.(*virtualfriends_go.VfRequest_StreamReplyMessage).StreamReplyMessage
			HandleStreamReplyMessage(handlingCtx, vfContext, request)

		case *virtualfriends_go.VfRequest_DownloadAssetBundle:
			err = fmt.Errorf("deprecated")
			logger.Error(err)
			vfContext.sendResp(FromError(err))
			return

		case *virtualfriends_go.VfRequest_DownloadBlob:
			request := vfRequest.Request.(*virtualfriends_go.VfRequest_DownloadBlob).DownloadBlob
			HandleDownloadBlob(handlingCtx, vfContext, request)

		case *virtualfriends_go.VfRequest_GetCharacter:
			request := vfRequest.Request.(*virtualfriends_go.VfRequest_GetCharacter).GetCharacter
			HandleGetCharacter(handlingCtx, vfContext, request)

		case *virtualfriends_go.VfRequest_AccumulateVoiceMessage:
			request := vfRequest.Request.(*virtualfriends_go.VfRequest_AccumulateVoiceMessage).AccumulateVoiceMessage
			HandleAccumulateVoiceMessage(handlingCtx, vfContext, request)
		}
	}
}
