package core

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sieglu2/virtual-friends-brain/foundation"
	"github.com/sieglu2/virtual-friends-brain/virtualfriends_go"
	"google.golang.org/protobuf/proto"
)

var upgrader = websocket.Upgrader{} // use default options

type VfContext struct {
	// use this delegate to send VfResponse to client single or multiple times.
	SendResp func(vfResponse *virtualfriends_go.VfResponse) error
}

func OnConnect(conn *websocket.Conn) *VfContext {
	logger := foundation.Logger()

	return &VfContext{
		SendResp: func(vfResponse *virtualfriends_go.VfResponse) error {
			vfResponseBytes, err := proto.Marshal(vfResponse)
			if err != nil {
				logger.Errorf("failed to marshal: %v", err)
				return err
			}
			return conn.WriteMessage(websocket.BinaryMessage, vfResponseBytes)
		},
	}
}

func OnDisconnect(vfContext *VfContext) {
	logger := foundation.Logger()
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
		mt, message, err := conn.ReadMessage()
		if err != nil {
			OnDisconnect(vfContext)
			break
		}

		vfRequest := &virtualfriends_go.VfRequest{}
		if err = proto.Unmarshal(message, vfRequest); err != nil {
			logger.Errorf("failed to unmarshal: %v", err)
			break
		}

		switch vfRequest.Request.(type) {
		case *virtualfriends_go.VfRequest_Echo:
			logger.Errorf("not supported for now")

		case *virtualfriends_go.VfRequest_StreamReplyMessage:

		case *virtualfriends_go.VfRequest_DownloadAssetBundle:
			logger.Errorf("deprecated")

		case *virtualfriends_go.VfRequest_DownloadBlob:

		case *virtualfriends_go.VfRequest_GetCharacter:
		}

		err = conn.WriteMessage(mt, message)
		if err != nil {
			logger.Errorf("failed to write: %v", err)
			break
		}
	}
}
