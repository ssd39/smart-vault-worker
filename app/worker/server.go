package worker

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ssd39/smart-vault-sgx-app/app/chainhelper"
)

var isSideCardConnected = false
var websocketConn *websocket.Conn
var upgrader = websocket.Upgrader{}
var mu sync.Mutex

var subReqList = []chainhelper.SubRequest{}

func SidecarChannel(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	if isSideCardConnected {
		defer mu.Unlock()
		jsonData, err := json.Marshal(FailedConnectRes{
			Sucess:  false,
			Message: "Sidecar already connected!",
		})
		if err != nil {
			logger.Error("Error marshaling JSON:", err)
			return
		}
		_, err = w.Write(jsonData)
		if err != nil {
			log.Println("Error writing JSON to WebSocket:", err)
		}
		return
	}
	isSideCardConnected = true
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("upgrade:", err)
		return
	}
	websocketConn = c
	mu.Unlock()
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		logger.Info("ws:message-type", mt)
		if err != nil {
			logger.Info("error while reading ws message:", err)
			mu.Lock()
			isSideCardConnected = false
			mu.Unlock()
			break

		}
		var data map[string]interface{}
		err = json.Unmarshal(message, &data)
		if err != nil {
			logger.Error("Error while decoding json", "err", string(message))
			continue
		}

		if msgType, ok := data["Type"]; ok {
			if msgType == "BidRes" {
				var bidRes BidRes
				err = json.Unmarshal(message, &bidRes)
				if err == nil && bidRes.IsApproved {
					ConfirmBid(&bidRes)
				}
			}
		}
	}
}

func ConfirmBid(bidRes *BidRes) {
	mu.Lock()
	defer mu.Unlock()
	var tempList []chainhelper.SubRequest
	now := time.Now()
	for _, item := range subReqList {
		if item.BidEndTime > now.Unix() {
			if bidRes.Id == item.Id {
				// call smart contrct to bid
			} else {
				tempList = append(tempList, item)
			}
		}
	}
	subReqList = tempList
}

func SendBidReq(subReqPayload chainhelper.SubRequest) error {
	mu.Lock()
	defer mu.Unlock()
	if !isSideCardConnected {
		return errors.New("Sidecar not connected!")
	}
	bidReqPayload := BidReq{Type: "BidReq", Id: subReqPayload.Id, MaxRent: subReqPayload.MaxRent}
	jsonData, err := json.Marshal(bidReqPayload)
	if err != nil {
		logger.Error("Error marshaling JSON:", "err", err)
		return err
	}
	err = websocketConn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		logger.Error("error while sending data on ws", "err", err)
		return err
	}
	subReqList = append(subReqList, subReqPayload)
	return nil
}

func SendExecuteReq() {

}
