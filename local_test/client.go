package main

import (
	"bufio"
	"flag"
	"log"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
	"github.com/sieglu2/virtual-friends-brain/foundation"
)

func ClientHandler(address string) {
	logger := foundation.Logger()

	// Setup the WebSocket URL
	u := url.URL{
		Scheme: "ws",
		Host:   address,
		Path:   "/echo",
	}
	logger.Infof("connecting to %s", u.String())

	// Connect to WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				logger.Warnf("Disconnecting...")
				break
			}
			logger.Infof("message: %s", string(message))
		}
	}()

	for {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			input := scanner.Text()
			logger.Infof("scanned: %s", input)

			if input == "<quit>" {
				conn.Close()
				break
			}

			err = conn.WriteMessage(websocket.TextMessage, []byte(input))
			if err != nil {
				logger.Errorf("failed to write:", err)
				return
			}
		}
	}
}

func main() {
	logger := foundation.Logger()

	addressPtr := flag.String("address", "", "remote address")
	flag.Parse()

	address := *addressPtr

	if len(address) == 0 {
		logger.Errorf("empty address!")
		return
	}

	ClientHandler(address)
}
