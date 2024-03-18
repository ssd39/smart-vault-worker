package test

import (
	"testing"

	"github.com/near/borsh-go"
	"github.com/ssd39/smart-vault-sgx-app/app/chainhelper"
)

func TestBorsh(t *testing.T) {
	initInstruction := chainhelper.InitData{Vault_public_key: [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}, Attestation_proof: "HelloSmartVault"}
	data, _ := borsh.Serialize(initInstruction)
	t.Log(data)
}
