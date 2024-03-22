package chainhelper

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
