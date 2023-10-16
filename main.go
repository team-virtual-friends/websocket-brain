package main

import (
	"flag"
	"net/http"

	"github.com/gorilla/websocket"
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
	flag.Parse()

	http.HandleFunc("/echo", echo)
	http.HandleFunc("/in-game", core.InGame)

	foundation.Logger().Fatal(http.ListenAndServe(*addr, nil))
}
