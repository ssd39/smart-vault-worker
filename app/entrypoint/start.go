package entrypoint

import (
	"net/http"

	"github.com/ssd39/smart-vault-sgx-app/app/chainhelper"
	"github.com/ssd39/smart-vault-sgx-app/app/worker"
)

func Start(keyPath string) error {
	/*var account types.Account
	if keyPath != "" {
		account = chainhelper.RecoverRootAccout(keyPath)
	} else {
		account = chainhelper.CreateAccount(true)
	}*/

	/*seed, err := smvCrypto.UnSealKey(".seedKey")
	if err != nil {
		logger.Error("Error while unselaing the seed key")
		return err
	}*/
	http.HandleFunc("/", worker.SidecarChannel)
	go http.ListenAndServe("localhost:8888", nil)

	eventChan := make(chan chainhelper.Instruction, 1000)
	err := chainhelper.ListenEvents(eventChan)
	if err != nil {
		return err
	}
	for event := range eventChan {

		logger.Info("new-event", "event", event)

		switch event.GetType() {
		case chainhelper.BidRequestType:
		case chainhelper.JoinReqType:
		case chainhelper.ProtocolInitType:
		case chainhelper.SubClosedType:
		case chainhelper.SubRequestType:
			if subReqPayload, ok := event.(chainhelper.SubRequest); ok {
				now, err := chainhelper.GetCurTime()
				if err != nil {
					eventChan <- event
					continue
				}
				if subReqPayload.BidEndTime < now {
					logger.Error("Got expired subreq", "payload", subReqPayload)
					continue
				}
				err = worker.SendBidReq(subReqPayload)
				if err != nil {
					eventChan <- event
					logger.Error("Error while sending subreq to sidecar", "action", "retrying")
				}
				logger.Info("Sent bidReq to sidecar")
			}
		}
	}
	return nil
}
