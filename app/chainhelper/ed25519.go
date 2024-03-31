package chainhelper

import (
	"errors"

	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/pkg/bincode"
	"github.com/blocto/solana-go-sdk/types"
)

var ed25519VerifyProgram = common.PublicKeyFromString("Ed25519SigVerify111111111111111111111111111")

const PRIVATE_KEY_BYTES = 64
const PUBLIC_KEY_BYTES = 32
const SIGNATURE_BYTES = 64

type ED25519InstructionLayout struct {
	NumSignatures             uint8
	Padding                   uint8
	SignatureOffset           uint16
	SignatureInstructionIndex uint16
	PublicKeyOffset           uint16
	PublicKeyInstructionIndex uint16
	MessageDataOffset         uint16
	MessageDataSize           uint16
	MessageInstructionIndex   uint16
}

func NewEd25519Instruction(pubKey []byte, signature []byte, message []byte) (types.Instruction, error) {
	if len(pubKey) != PUBLIC_KEY_BYTES {
		return types.Instruction{}, errors.New("Invlaid pubkey")
	}
	if len(signature) != SIGNATURE_BYTES {
		return types.Instruction{}, errors.New("Invlaid signature")
	}
	publicKeyOffset := 16
	signatureOffset := publicKeyOffset + len(pubKey)
	messageDataOffset := signatureOffset + len(signature)
	numSignatures := 1
	instructionData := make([]byte, messageDataOffset+len(message))
	index := uint16(0xffff)
	instructionLayout := ED25519InstructionLayout{NumSignatures: uint8(numSignatures), SignatureOffset: uint16(signatureOffset), SignatureInstructionIndex: index, PublicKeyOffset: uint16(publicKeyOffset), PublicKeyInstructionIndex: index, MessageDataOffset: uint16(messageDataOffset), MessageDataSize: uint16(len(message)), MessageInstructionIndex: index}
	instructionLayoutBytes, err := bincode.SerializeData(instructionLayout)
	if err != nil {
		logger.Error("Error whiel serialise layout data")
		return types.Instruction{}, err
	}
	if len(instructionLayoutBytes) != publicKeyOffset {
		logger.Error("Data not serialised properly!")
		return types.Instruction{}, errors.New("serialised not properly!")
	}
	copy(instructionData[:], instructionLayoutBytes)
	copy(instructionData[publicKeyOffset:], pubKey)
	copy(instructionData[signatureOffset:], signature)
	copy(instructionData[messageDataOffset:], message)

	return types.Instruction{
		ProgramID: ed25519VerifyProgram,
		Data:      instructionData,
	}, nil
}
