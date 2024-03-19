package chainhelper

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"log"
	"os"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/types"
	"github.com/gorilla/websocket"
	"github.com/mr-tron/base58"
	"github.com/near/borsh-go"

	"github.com/ssd39/smart-vault-sgx-app/app/utils"
)

var RpcEndpoint = "http://solana-rpc.oraculus.network"
var WsRpcEndpoint = "ws://127.0.0.1:8900"
var programId = common.PublicKeyFromString("DmrxZbSZF84Pa9dnrg23FsvZffZpovfgi5pjqDKoDxom")
var systemProgramm = common.PublicKeyFromString("11111111111111111111111111111111")

var VAULT_METADATA = "METADATA"
var APP_COUNTER = "APP_COUNTER"

func init() {
	logger = utils.Logger
}

func CreateAccount(rootAcc bool) types.Account {
	account := types.NewAccount()
	if rootAcc {

		logger.Info("Creating new keypair!")

		dir, err := os.Getwd() //get the current directory using the built-in function
		if err != nil {
			logger.Error("Failed to get cwd!", err)
			os.Exit(1)
		}
		logger.Info("Keypair:" + dir + "/.walletKey")
		err = os.WriteFile("./vault/.walletKey", []byte(base58.Encode(account.PrivateKey)), 0644)
		if err != nil {
			logger.Error("Failed to create .walletKey file!", err)
			os.Exit(1)
		}
	}

	return account
}

func RecoverAccount(key string) (types.Account, error) {
	return types.AccountFromBase58(key)
}

func RecoverAccountFromPK(priKey ed25519.PrivateKey) types.Account {
	return types.Account{
		PublicKey:  common.PublicKeyFromBytes(priKey.Public().(ed25519.PublicKey)),
		PrivateKey: priKey,
	}
}

func RecoverRootAccout(keyPath string) types.Account {
	dat, err := os.ReadFile(keyPath)
	if err != nil {
		logger.Error("Failed to read walletkey", err)
		os.Exit(1)
	}
	acc, err := RecoverAccount(string(dat))
	if err != nil {
		logger.Error("Failed to recover walletkey", err)
		os.Exit(1)
	}
	return acc
}

func Join(account types.Account, consesues types.Account, attestation string) (string, error) {
	c := client.NewClient(RpcEndpoint)
	res, err := c.GetLatestBlockhash(context.Background())
	if err != nil {
		logger.Error("Error while getting latest block")
		return "", err
	}

	initInstruction := InitData{Vault_public_key: [32]byte(consesues.PublicKey.Bytes()), Attestation_proof: attestation}
	data, err := borsh.Serialize(initInstruction)
	if err != nil {
		logger.Errorf("Error while borsh seralisation")
		return "", err
	}

	vaultStrBytes := []byte(VAULT_METADATA)
	var seeds [][]byte
	seeds = append(seeds, vaultStrBytes)
	vaultMetaDataPda, _, err := common.FindProgramAddress(seeds, programId)

	if err != nil {
		logger.Error("Failed to derive pda account for vault metadata")
		return "", err
	}

	counterStrBytes := []byte(APP_COUNTER)
	var counterSeeds [][]byte
	counterSeeds = append(counterSeeds, counterStrBytes)
	counterPda, _, err := common.FindProgramAddress(counterSeeds, programId)
	if err != nil {
		logger.Error("Failed to derive pda account for counter")
		return "", err
	}

	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{account},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        account.PublicKey,
			RecentBlockhash: res.Blockhash,
			Instructions: []types.Instruction{
				{
					ProgramID: programId,
					Accounts: []types.AccountMeta{
						{
							PubKey:   account.PublicKey,
							IsSigner: true,
						}, {
							PubKey:     vaultMetaDataPda,
							IsWritable: true,
						},
						{
							PubKey:     counterPda,
							IsWritable: true,
						},
						{
							PubKey: systemProgramm,
						},
					},
					Data: utils.Prepend([]byte{0}, data),
				},
			},
		}),
	})

	if err != nil {
		logger.Errorf("failed to new a tx, err: %v", err)
		return "", err
	}

	sig, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		logger.Errorf("failed to send the tx, err: %v", err)
		return "", err
	}

	return sig, nil
}

type Message struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

func ListenEvents() error {
	conn, _, err := websocket.DefaultDialer.Dial(WsRpcEndpoint, nil)
	if err != nil {
		logger.Error("Error connecting web-socket")
		return err
	}
	defer conn.Close()

	subscribeMessage := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "logsSubscribe",
		"params": []interface{}{
			map[string]interface{}{
				"mentions": []string{programId.ToBase58()},
			},
			map[string]interface{}{
				"commitment": "finalized",
			},
		},
	}
	err = conn.WriteJSON(subscribeMessage)
	if err != nil {
		logger.Error("Error subscribing to account")
		return err
	}
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Fatal("Error reading message from Solana RPC:", err)
		}

		// Unmarshal message
		var message Message
		err = json.Unmarshal(msg, &message)
		if err != nil {
			logger.Error("Error unmarshalling message", "error", err)
			continue
		}

		// Handle message
		switch message.Method {
		case "logsNotification":
			logger.Info("Received account notification", "msg", string(message.Params))
		default:
			logger.Info("Unhandled message", "msg", string(msg))
		}
	}
}
