package chainhelper

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/json"
	"log"
	"os"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/program/associated_token_account"
	"github.com/blocto/solana-go-sdk/types"
	"github.com/gorilla/websocket"
	"github.com/mr-tron/base58"
	"github.com/near/borsh-go"

	"github.com/ssd39/smart-vault-sgx-app/app/utils"
)

var RpcEndpoint = "http://solana-rpc.oraculus.network"
var WsRpcEndpoint = "ws://127.0.0.1:8900"
var programId = common.PublicKeyFromString("6bcSZLTvfu2ZaC7yhXfkaupFG315r4qWK8wqSQN5LRFT")
var systemProgram = common.PublicKeyFromString("11111111111111111111111111111111")
var smvSplToken = common.PublicKeyFromString("5DYw4t2nJoSyhD9NDnPTveN7ZY4DwZyDXHTMJPdnqeZG")

var VAULT_METADATA = "METADATA"
var APP_COUNTER = "APP_COUNTER"
var TREASURY_STATE = "TREASURY_STATE"

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

func GetCurTime() (int64, error) {
	c := client.NewClient(RpcEndpoint)
	clock, err := c.GetAccountInfo(context.Background(), common.SysVarClockPubkey.ToBase58())
	if err != nil {
		logger.Error(err)
		return 0, err
	}
	var clockData ClockData
	reader := bytes.NewReader(clock.Data)
	err = binary.Read(reader, binary.LittleEndian, &clockData)
	if err != nil {
		return 0, err
	}
	return clockData.UnixTimestamp, nil
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

	tresuaryStrBytes := []byte(TREASURY_STATE)
	var tresuarySeeds [][]byte
	tresuarySeeds = append(tresuarySeeds, tresuaryStrBytes)
	tresuaryPda, _, err := common.FindProgramAddress(tresuarySeeds, programId)
	if err != nil {
		logger.Error("Failed to derive pda account for counter")
		return "", err
	}

	programAta, _, err := common.FindAssociatedTokenAddress(tresuaryPda, smvSplToken)
	if err != nil {
		logger.Error("Failed to derive pda account for program ata")
		return "", err
	}

	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{account},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        account.PublicKey,
			RecentBlockhash: res.Blockhash,
			Instructions: []types.Instruction{
				associated_token_account.Create(associated_token_account.CreateParam{
					Funder:                 account.PublicKey,
					Owner:                  tresuaryPda,
					Mint:                   smvSplToken,
					AssociatedTokenAccount: programAta,
				}),
				{
					ProgramID: programId,
					Accounts: []types.AccountMeta{
						{
							PubKey:   account.PublicKey,
							IsSigner: true,
						},
						{
							PubKey:     vaultMetaDataPda,
							IsWritable: true,
						},
						{
							PubKey:     counterPda,
							IsWritable: true,
						},
						{
							PubKey:     tresuaryPda,
							IsWritable: true,
						},
						{
							PubKey: systemProgram,
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

func ListenEvents(eventChan chan Instruction) error {
	conn, _, err := websocket.DefaultDialer.Dial(WsRpcEndpoint, nil)
	if err != nil {
		logger.Error("Error connecting web-socket")
		return err
	}

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
	go func() {
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
				logger.Info("Received logs notification", "msg", string(message.Params))
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(string(message.Params)), &data); err != nil {
					logger.Error("Error decoding JSON", "error", err)
					continue
				}

				logs, ok := data["result"].(map[string]interface{})["value"].(map[string]interface{})["logs"].([]interface{})
				if !ok {
					logger.Error("Error: Logs not found in JSON")
					continue
				}

				for _, log := range logs {
					instruction, err := ParseProgramLog(log.(string))
					if err == nil {
						eventChan <- instruction
					}
				}
			default:
				logger.Info("Unhandled message", "msg", string(msg))
			}
		}
	}()
	return nil
}
