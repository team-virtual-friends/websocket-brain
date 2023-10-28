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
			if trimmed == VirtualFriendsOrigin || strings.HasPrefix(trimmed, "http://localhost:") {
				return true
			}
		}
		return false
	},
} // use default options

type VfContext struct {
	// use this delegate to send VfResponse to client single or multiple times.
	sendResp func(vfResponse *virtualfriends_go.VfResponse) error

	// the reference to the VfRequest being processed, if it's outside the processing
	// handler it'll be the last one.
	originalVfRequest *virtualfriends_go.VfRequest

	remoteAddr string
	clients    *common.Clients

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

		remoteAddr: conn.RemoteAddr().String(),
		clients:    common.GetGlobalClients(),
	}
}

func OnDisconnect(vfContext *VfContext) {
	logger := foundation.Logger()

	if foundation.IsProd() {
		go logChatHistory(vfContext)
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

		vfContext.originalVfRequest = vfRequest

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
