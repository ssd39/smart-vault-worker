package worker

import "github.com/blocto/solana-go-sdk/common"

type FailedConnectRes struct {
	Sucess  bool
	Message string
}

type BidRes struct {
	Type       string `default:"BidRes"`
	Id         uint64
	Rent       uint64
	IsApproved bool
}

type BidReq struct {
	Type    string `default:"BidReq"`
	Id      uint64
	MaxRent uint64
}

type ExecuteReq struct {
	Id             uint64
	AppIpfsHash    string
	ParamsIpfsHash string
	User           common.PublicKey
}
