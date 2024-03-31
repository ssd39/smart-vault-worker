package sidecar

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/ssd39/smart-vault-sgx-app/app/worker"
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
		var data map[string]interface{}
		err = json.Unmarshal(message, &data)
		if err != nil {
			logger.Error("Error while decoding json", "err", string(message))
			continue
		}
		reqType, ok := data["Type"]
		if ok {
			if reqType == "BidReq" {
				var bidReq worker.BidReq
				err = json.Unmarshal(message, &bidReq)
				if err != nil {
					logger.Error("Error while un-marshaling")
				}
				var bidRes worker.BidRes
				bidRes.Type = "BidRes"
				bidRes.Id = bidReq.Id
				bidRes.IsApproved = true
				bidRes.Rent = bidReq.MaxRent
				jsonData, err := json.Marshal(bidRes)
				if err != nil {
					logger.Error("Error marshaling JSON:", "err", err)
					continue
				}
				c.WriteMessage(websocket.TextMessage, jsonData)
			}
		}
	}
}
