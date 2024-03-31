package worker

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/blocto/solana-go-sdk/types"
	"github.com/gorilla/websocket"
	"github.com/ssd39/smart-vault-sgx-app/app/chainhelper"
)

var isSideCardConnected = false
var websocketConn *websocket.Conn
var upgrader = websocket.Upgrader{}
var mu sync.Mutex

var subReqList = []chainhelper.SubRequest{}
var subMap = map[uint64]bool{}
var account types.Account
var concesuesAcc types.Account

func SetAccount(acc types.Account) {
	account = acc
}

func SetConsesuesAccount(acc types.Account) {
	concesuesAcc = acc
}

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
	now, err := chainhelper.GetCurTime()
	if err != nil {
		return
	}
	for _, item := range subReqList {
		if item.BidEndTime > now {
			if bidRes.Id == item.Id {
				// call smart contrct to bid
				_, err := chainhelper.AddBid(account, concesuesAcc, item, bidRes.Rent)
				if err == nil {
					curTime, err := chainhelper.GetCurTime()
					if err == nil {
						go func(item_ chainhelper.SubRequest) {
							logger.Info("Claim can be proceed after", "seconds", item_.BidEndTime-curTime)
							timer := time.NewTimer(time.Duration(item_.BidEndTime-curTime) * time.Second)
							<-timer.C
							timer.Stop()
							// check if i am the bid winner if it is claim the bid
							chainhelper.ClaimBid(account, concesuesAcc, item_)
						}(item)
					}
				}
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
	_, isExsist := subMap[subReqPayload.Id]
	if isExsist {
		return nil
	}
	subMap[subReqPayload.Id] = true
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
