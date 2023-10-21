package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sieglu2/virtual-friends-brain/common"
	"github.com/sieglu2/virtual-friends-brain/core"
	"github.com/sieglu2/virtual-friends-brain/foundation"
)

var addr = flag.String("addr", "localhost:8510", "Virtual Friends Brain")

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	logger := foundation.Logger()

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("failed to upgrade: %v", err)
		return
	}
	defer c.Close()

	logger.Infof("connected from: %+v", c.RemoteAddr())

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			// logger.Errorf("failed to read: %v", err)
			logger.Infof("disconnected to %+v", c.RemoteAddr())
			break
		}

		logger.Infof("recv: %s", message)

		err = c.WriteMessage(mt, message)
		if err != nil {
			logger.Errorf("failed to write: %v", err)
			break
		}
	}
}

func main() {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "./secrets/ysong-chat-845e43a6c55b.json")

	flag.Parse()

	logger := foundation.Logger()
	initCtx, initCancel := context.WithTimeout(context.Background(), 60*time.Second)

	if err := common.InitializeClients(initCtx); err != nil {
		logger.Fatalf("failed to initialize clients: %v", err)
	}

	if err := core.DownloadAllAssetBundles(initCtx); err != nil {
		logger.Fatalf("failed to download all assetbundles: %v", err)
	}

	initCancel()

	http.HandleFunc("/echo", echo)
	http.HandleFunc("/in-game", core.InGame)

	logger.Infof("starting server...")
	logger.Fatal(http.ListenAndServe(*addr, nil))
}
