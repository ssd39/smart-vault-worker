package entrypoint

import (
	"crypto/ed25519"

	"github.com/blocto/solana-go-sdk/types"
	"github.com/edgelesssys/ego/enclave"
	"github.com/ssd39/smart-vault-sgx-app/app/chainhelper"
	smvCrypto "github.com/ssd39/smart-vault-sgx-app/app/crypto"
	"github.com/ssd39/smart-vault-sgx-app/app/ipfs"
	"github.com/ssd39/smart-vault-sgx-app/app/utils"
)

func Init(keyPath string, ipfsUploader ipfs.IpfsUploader) error {
	var account types.Account
	if keyPath != "" {
		account = chainhelper.RecoverRootAccout(keyPath)
	} else {
		account = chainhelper.CreateAccount(true)
	}
	logger.Info(account.PublicKey.ToBase58())
	seed := utils.GenerateRandom64Bytes()[:32]
	privKey := ed25519.NewKeyFromSeed(seed)
	// temp logic to test things on non-sgx env
	logger.Info("Seed:", "=>", privKey.Seed())
	err := smvCrypto.SealKeyToFile(privKey.Seed(), ".seedKey")
	if err != nil {
		return err
	}

	concesuesAcc := chainhelper.RecoverAccountFromPK(privKey)

	signedMessage := concesuesAcc.Sign(concesuesAcc.PublicKey.Bytes())

	report, err := enclave.GetRemoteReport(signedMessage)
	if err != nil {
		return err
	}

	cid, err := ipfs.UploadBytes(ipfsUploader, report)
	if err != nil {
		logger.Error("Failed to upload attestation report on ipfs")
		return err
	}
	logger.Infof("CID: %s", cid)

	sig, err := chainhelper.Join(account, concesuesAcc, cid)
	if err != nil {
		return err
	}
	logger.Infof("Protocol Init Tx: %s", sig)
	return nil
}
