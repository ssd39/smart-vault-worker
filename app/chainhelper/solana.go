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
var programId = common.PublicKeyFromString("J82e6dQCfCgNdfQmyAdAX9jFeE7cREE84gfT3ViNauDN")
var systemProgram = common.PublicKeyFromString("11111111111111111111111111111111")
var smvSplToken = common.PublicKeyFromString("F9uNojiqaWU8FPtiBhbmZtewsWFCCGk38eRDRqUTxg7L")

var VAULT_METADATA = "METADATA"
var APP_COUNTER = "APP_COUNTER"
var TREASURY_STATE = "TREASURY_STATE"
var BIDDER_STATE = "BIDDER_STATE"
var SUB_STATE = "SUB_STATE"

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

func AddBid(account types.Account, consesues types.Account, subReq SubRequest, bidAmount uint64) (string, error) {
	c := client.NewClient(RpcEndpoint)
	res, err := c.GetLatestBlockhash(context.Background())
	if err != nil {
		logger.Error("Error while getting latest block")
		return "", err
	}

	sigProveData, err := getConsesuesSigProoveForBidder(c, account, consesues)
	if err != nil {
		return "", err
	}
	addBidData := AddBidData{Signature: sigProveData.Signature, BidAmount: bidAmount}
	data, err := borsh.Serialize(addBidData)
	if err != nil {
		logger.Errorf("Error while borsh seralisation")
		return "", err
	}

	subIdBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(subIdBytes, subReq.Id)

	var subStateSeeds [][]byte
	subStrBytes := []byte(SUB_STATE)
	subStateSeeds = append(subStateSeeds, subStrBytes)
	subStateSeeds = append(subStateSeeds, subReq.SubScriberKey.Bytes())
	subStateSeeds = append(subStateSeeds, subIdBytes)
	subStatePda, _, err := common.FindProgramAddress(subStateSeeds, programId)
	if err != nil {
		return "", err
	}
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{account},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        account.PublicKey,
			RecentBlockhash: res.Blockhash,
			Instructions: []types.Instruction{
				sigProveData.Ed2519Instruction,
				{
					ProgramID: programId,
					Accounts: []types.AccountMeta{
						{
							PubKey: consesues.PublicKey,
						},
						{
							PubKey:   account.PublicKey,
							IsSigner: true,
						},
						{
							PubKey:     sigProveData.BidderStatePda,
							IsWritable: true,
						},
						{
							PubKey:     subStatePda,
							IsWritable: true,
						},
						{
							PubKey: sigProveData.VaultMetaDataPda,
						},
						{
							PubKey: systemProgram,
						},
						{
							PubKey: common.SysVarInstructionsPubkey,
						},
					},
					Data: utils.Prepend([]byte{5}, data),
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
	logger.Info("AddBid", "signature", sig)
	return sig, nil
}

func ClaimBid(account types.Account, consesues types.Account, subReq SubRequest) (string, error) {
	logger.Info("Trying to claim bid!")
	c := client.NewClient(RpcEndpoint)

	subIdBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(subIdBytes, subReq.Id)

	var subStateSeeds [][]byte
	subStrBytes := []byte(SUB_STATE)
	subStateSeeds = append(subStateSeeds, subStrBytes)
	subStateSeeds = append(subStateSeeds, subReq.SubScriberKey.Bytes())
	subStateSeeds = append(subStateSeeds, subIdBytes)
	subStatePda, _, err := common.FindProgramAddress(subStateSeeds, programId)
	if err != nil {
		logger.Info("Failed to get substate pda", "error", err)
		return "", err
	}
	subStateAcc, err := c.GetAccountInfo(context.Background(), subStatePda.ToBase58())
	if err != nil {
		logger.Info("Failed to get acc substate info", "error", err)
		return "", err
	}

	var subStateData VaultUserSubscriptionStateData
	err = borsh.Deserialize(&subStateData, subStateAcc.Data)
	if err != nil {
		logger.Info("Failed to deserailise substate data", "error", err)
		return "", nil
	}
	logger.Info("Subscription state", "=", subStateData)
	if subStateData.Executor == account.PublicKey {
		logger.Info("I wont the bid processing further to claim it!")
		// I am the winner lets claim bid
		res, err := c.GetLatestBlockhash(context.Background())
		if err != nil {
			logger.Error("Error while getting latest block")
			return "", err
		}
		sigProveData, err := getConsesuesSigProoveForBidder(c, account, consesues)
		if err != nil {
			return "", err
		}
		claimBid := ClaimBidData{Signature: sigProveData.Signature}
		data, err := borsh.Serialize(claimBid)
		if err != nil {
			logger.Errorf("Error while borsh seralisation")
			return "", err
		}
		tx, err := types.NewTransaction(types.NewTransactionParam{
			Signers: []types.Account{account},
			Message: types.NewMessage(types.NewMessageParam{
				FeePayer:        account.PublicKey,
				RecentBlockhash: res.Blockhash,
				Instructions: []types.Instruction{
					sigProveData.Ed2519Instruction,
					{
						ProgramID: programId,
						Accounts: []types.AccountMeta{
							{
								PubKey: consesues.PublicKey,
							},
							{
								PubKey:   account.PublicKey,
								IsSigner: true,
							},
							{
								PubKey:     sigProveData.BidderStatePda,
								IsWritable: true,
							},
							{
								PubKey:     subStatePda,
								IsWritable: true,
							},
							{
								PubKey: sigProveData.VaultMetaDataPda,
							},
							{
								PubKey: common.SysVarInstructionsPubkey,
							},
						},
						Data: utils.Prepend([]byte{6}, data),
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
		logger.Info("ClaimBid", "signature", sig)
		return sig, nil
	}
	return "", nil
}

func getConsesuesSigProoveForBidder(c *client.Client, account types.Account, consesues types.Account) (SigProveData, error) {
	vaultStrBytes := []byte(VAULT_METADATA)
	var seeds [][]byte
	seeds = append(seeds, vaultStrBytes)
	vaultMetaDataPda, _, err := common.FindProgramAddress(seeds, programId)
	if err != nil {
		logger.Error("Failed to derive pda account for vault metadata")
		return SigProveData{}, err
	}

	var bidderSeeds [][]byte
	bidderSeeds = append(bidderSeeds, []byte(BIDDER_STATE))
	bidderSeeds = append(bidderSeeds, account.PublicKey.Bytes())
	bidderStatePda, _, err := common.FindProgramAddress(bidderSeeds, programId)
	if err != nil {
		return SigProveData{}, err
	}
	bidderAcc, err := c.GetAccountInfo(context.Background(), bidderStatePda.ToBase58())
	if err != nil {
		return SigProveData{}, err
	}

	bidderNonce := uint64(0)
	if len(bidderAcc.Data) > 0 {
		var bidderState BidderStateData
		err = borsh.Deserialize(&bidderState, bidderAcc.Data)
		if err != nil {
			logger.Info("Error while deserializing data")
			return SigProveData{}, err
		}
		bidderNonce = bidderState.Nonce
	}

	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, bidderNonce)
	var rawMsg [][]byte
	rawMsg = append(rawMsg, account.PublicKey.Bytes())
	rawMsg = append(rawMsg, nonceBytes)

	signature := consesues.Sign(bytes.Join(rawMsg, nil))
	ed2519Instruction, err := NewEd25519Instruction(consesues.PublicKey.Bytes(), signature, bytes.Join(rawMsg, nil))
	if err != nil {
		logger.Info("Error while creating ed2519Instruction")
		return SigProveData{}, err
	}
	return SigProveData{Ed2519Instruction: ed2519Instruction, BidderStatePda: bidderStatePda, VaultMetaDataPda: vaultMetaDataPda, Signature: [64]byte(signature)}, nil
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
