package entrypoint

import (
	"crypto/ed25519"

	"github.com/blocto/solana-go-sdk/types"
	"github.com/ssd39/smart-vault-sgx-app/app/chainhelper"
	smvCrypto "github.com/ssd39/smart-vault-sgx-app/app/crypto"
	"github.com/ssd39/smart-vault-sgx-app/app/utils"
)

var AttestationMessage = "HelloSmartVault"

func Init(keyPath string) error {
	var account types.Account
	if keyPath != "" {
		account = chainhelper.RecoverRootAccout(keyPath)
	} else {
		account = chainhelper.CreateAccount(true)
	}
	logger.Info(account.PublicKey.ToBase58())
	seed := utils.GenerateRandom64Bytes()[:32]
	privKey := ed25519.NewKeyFromSeed(seed)
	err := smvCrypto.SealKeyToFile(privKey.Seed(), ".seedKey")
	if err != nil {
		return err
	}

	concesuesAcc := chainhelper.RecoverAccountFromPK(privKey)

	//signedMessage := concesuesAcc.Sign([]byte(AttestationMessage))

	// TODO: storing report to ipfs
	/*report, err := enclave.GetRemoteReport(signedMessage)
	if err != nil {
		return err
	}*/

	// for test only
	//logger.Info(report)
	sig, err := chainhelper.Join(account, concesuesAcc, string([]byte(AttestationMessage)))
	if err != nil {
		return err
	}
	logger.Infof("Protocol Init Tx: %s", sig)
	return nil
}
