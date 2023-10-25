package core

import (
	"context"
	"encoding/json"
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
		origins := r.Header["Origin"]
		if len(origins) == 0 {
			return false
		}
		for _, origin := range origins {
			if origin == VirtualFriendsOrigin {
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
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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

	go logChatHistory(vfContext)
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
			err = fmt.Errorf("not supported for now")
			logger.Error(err)
			vfContext.sendResp(FromError(err))
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
		}
	}
}

func logChatHistory(vfContext *VfContext) error {
	logger := foundation.Logger()

	vfRequest := vfContext.originalVfRequest
	if vfRequest == nil {
		logger.Warnf("nil vfRequest in vfContext")
		return nil
	}
	request := vfRequest.Request

	var characterId, chatHistory string
	switch request.(type) {
	case *virtualfriends_go.VfRequest_StreamReplyMessage:
		streamReplyMessageRequest := request.(*virtualfriends_go.VfRequest_StreamReplyMessage).StreamReplyMessage
		characterId = streamReplyMessageRequest.MirroredContent.CharacterId
		jsonMessages := streamReplyMessageRequest.JsonMessages
		chatHistory = assembleChatHistory(jsonMessages)
	}

	if len(chatHistory) > 0 {
		bqClient := vfContext.clients.GetBigQueryClient()
		err := bqClient.WriteChatHistory(&common.ChatHistory{
			UserId:        vfRequest.UserId,
			UserIp:        vfContext.remoteAddr,
			CharacterId:   characterId,
			ChatHistory:   chatHistory,
			Timestamp:     time.Now(),
			ChatSessionId: vfContext.originalVfRequest.SessionId,
			RuntimeEnv:    vfContext.originalVfRequest.RuntimeEnv.String(),
		})
		if err != nil {
			err = fmt.Errorf("failed to WriteChatHistory: %v", err)
			logger.Error(err)
			return err
		}
	}
	logger.Info("done writing the chat history")
	return nil
}

func assembleChatHistory(jsonMessages []string) string {
	resultBuilder := strings.Builder{}

	currentRole := ""
	combinedContent := strings.Builder{}
	for _, jsonMessage := range jsonMessages {
		if len(jsonMessage) == 0 {
			continue
		}

		var message chatMessage
		err := json.Unmarshal([]byte(jsonMessage), &message)
		if err != nil {
			resultBuilder.WriteString("---<missing>---")
			continue
		}

		if len(currentRole) > 0 && currentRole != message.Role {
			if currentRole == "assistant" {
				resultBuilder.WriteString("A")
			} else {
				resultBuilder.WriteString(currentRole)
			}

			resultBuilder.WriteString(": ")
			resultBuilder.WriteString(strings.Trim(combinedContent.String(), " "))
			combinedContent.Reset()
		}
		separater := ""
		if combinedContent.Len() > 0 {
			separater = ". "
		}
		combinedContent.WriteString(separater)
		combinedContent.WriteString(message.Content)
		currentRole = message.Role
	}

	if combinedContent.Len() > 0 {
		if currentRole == "assistant" {
			resultBuilder.WriteString("A")
		} else {
			resultBuilder.WriteString(currentRole)
		}

		resultBuilder.WriteString(": ")
		resultBuilder.WriteString(combinedContent.String())
	}

	return resultBuilder.String()
}
