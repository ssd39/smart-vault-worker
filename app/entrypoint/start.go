package entrypoint

import (
	"crypto/ed25519"
	"net/http"

	"github.com/ssd39/smart-vault-sgx-app/app/chainhelper"
	"github.com/ssd39/smart-vault-sgx-app/app/worker"
)

func Start(keyPath string) error {
	account := chainhelper.RecoverRootAccout(keyPath)
	worker.SetAccount(account)

	/*seed, err := smvCrypto.UnSealKey(".seedKey")
	if err != nil {
		logger.Error("Error while unselaing the seed key")
		return err
	}*/
	privKey := ed25519.NewKeyFromSeed([]byte{234, 49, 105, 86, 61, 230, 127, 221, 58, 99, 248, 209, 213, 52, 168, 83, 180, 181, 71, 80, 30, 228, 205, 188, 68, 217, 186, 51, 75, 82, 25, 198})
	// temp logic to test things on non-sgx env
	/*logger.Info("Seed:", "=>", privKey.Seed())
	err := smvCrypto.SealKeyToFile(privKey.Seed(), ".seedKey")
	if err != nil {
		return err
	}*/

	concesuesAcc := chainhelper.RecoverAccountFromPK(privKey)
	worker.SetConsesuesAccount(concesuesAcc)

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
