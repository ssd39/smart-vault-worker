package sidecar

import (
	"github.com/gorilla/websocket"
)

func StartListner() {
	c, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8888", nil)
	if err != nil {
		logger.Fatal("dial:", err)
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			logger.Info("read:", err)
			return
		}
		logger.Infof("recv: %s", message)
	}
}
