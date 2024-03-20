package entrypoint

import "github.com/ssd39/smart-vault-sgx-app/app/chainhelper"

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
	eventChan := make(chan chainhelper.Instruction)
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
		}
	}
	return nil
}
