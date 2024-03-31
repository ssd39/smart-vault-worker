package chainhelper

import (
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/types"
)

type InitData struct {
	Vault_public_key  [32]uint8
	Attestation_proof string
}

type ClockData struct {
	Slot                uint64
	EpochStartTimestamp int64
	Epoch               uint64
	LeaderScheduleEpoch uint64
	UnixTimestamp       int64
}

type BidderStateData struct {
	Is_initialized bool
	Nonce          uint64
}

type VaultUserSubscriptionStateData struct {
	ID             uint64
	IsInitialized  bool
	Closed         bool
	AppID          uint64
	ParamsHash     string
	MaxRent        uint64
	IsAssigned     bool
	Executor       [32]uint8
	BidEndTime     uint64
	Rent           uint64
	Nonce          uint64
	LastReportTime uint64
	Restart        bool
}

type SigProveData struct {
	Ed2519Instruction types.Instruction
	BidderStatePda    common.PublicKey
	VaultMetaDataPda  common.PublicKey
	Signature         [64]byte
}

type AddBidData struct {
	Signature [64]byte
	BidAmount uint64
}

type ClaimBidData struct {
	Signature [64]byte
}
