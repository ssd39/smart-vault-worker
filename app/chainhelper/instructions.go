package chainhelper

import (
	"errors"
	"strconv"
	"strings"

	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/types"
)

const JoinReqType = 0
const ProtocolInitType = 1
const SubRequestType = 2
const BidRequestType = 3
const SubClosedType = 4

type Instruction interface {
	GetType() uint8
}

type ProtocolInit struct {
	VaultKey    types.Account
	Attestation string
}

type JoinReq struct {
	TransitKey  common.PublicKey
	Attestation string
}

type SubRequest struct {
	SubScriberKey common.PublicKey
	Id            uint64
	MaxRent       uint64
	BidEndTime    uint64
}

type BidAdded struct {
	Bidder common.PublicKey
	Rent   uint64
}

type SubClosed struct {
	SubKey common.PublicKey
}

func (i ProtocolInit) GetType() uint8 {
	return ProtocolInitType
}

func (i JoinReq) GetType() uint8 {
	return JoinReqType
}

func (i SubRequest) GetType() uint8 {
	return SubRequestType
}

func (i BidAdded) GetType() uint8 {
	return BidRequestType
}

func (i SubClosed) GetType() uint8 {
	return SubClosedType
}

func ParseProgramLog(log string) (Instruction, error) {
	logParts := strings.Split(log, ":")
	if len(logParts) >= 2 {
		logParts[1] = strings.Trim(logParts[1], " ")
		if logParts[1] == "ProtocolInit" {
			if len(logParts[1:]) == 3 {
				acc, err := types.AccountFromBase58(logParts[2])
				if err != nil {
					return nil, err
				}
				return ProtocolInit{
					VaultKey:    acc,
					Attestation: logParts[3],
				}, nil
			}
		} else if logParts[1] == "JoinReq" {
			if len(logParts[1:]) == 3 {
				acc := common.PublicKeyFromString(logParts[2])
				return JoinReq{
					TransitKey:  acc,
					Attestation: logParts[3],
				}, nil
			}
		} else if logParts[1] == "SubRequest" {
			if len(logParts[1:]) == 5 {
				acc := common.PublicKeyFromString(logParts[2])

				id, err := strconv.ParseUint(logParts[3], 10, 64)
				if err != nil {
					return nil, err
				}
				maxRent, err := strconv.ParseUint(logParts[4], 10, 64)
				if err != nil {
					return nil, err
				}
				bidEndTime, err := strconv.ParseUint(logParts[5], 10, 64)
				if err != nil {
					return nil, err
				}
				return SubRequest{
					SubScriberKey: acc,
					Id:            id,
					MaxRent:       maxRent,
					BidEndTime:    bidEndTime,
				}, nil
			}
		} else if logParts[1] == "BidAdded" {
			if len(logParts[1:]) == 3 {
				acc := common.PublicKeyFromString(logParts[2])

				rent, err := strconv.ParseUint(logParts[3], 10, 64)
				if err != nil {
					return nil, err
				}
				return BidAdded{
					Bidder: acc,
					Rent:   rent,
				}, nil
			}
		} else if logParts[1] == "SubClosed" {
			if len(logParts[1:]) == 2 {
				subKey := common.PublicKeyFromString(logParts[2])
				return SubClosed{
					SubKey: subKey,
				}, nil
			}
		}
	}
	return nil, errors.New("Unsported log format")
}
