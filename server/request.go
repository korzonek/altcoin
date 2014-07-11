package server

import (
	"github.com/toqueteos/altcoin/config"
	"github.com/toqueteos/altcoin/tools"
	"github.com/toqueteos/altcoin/types"
)

type Request struct {
	// Required for all
	Version string `json:"version,omitempty"`
	Type    string `json:"type,omitempty"`
	// RangeRequest
	Range []int `json:"range,omitempty"`
	// PushTx
	*types.Tx `json:"tx,omitempty"`
	// PushBlock
	*types.Block `json:"block,omitempty"`
}

type Response struct {
	// SecurityCheck
	Secure bool   `json:"secure,omitempty"`
	Error  string `json:"error,omitempty"`
	// BlockCount
	Length     int    `json:"length,omitempty"`
	RecentHash int    `json:"recentHash,omitempty"`
	DiffLength string `json:"diffLength,omitempty"`
	// RangeRequest
	Blocks []*types.Block `json:"blocks,omitempty"`
	// Txs
	Txs []*types.Tx `json:"txs,omitempty"`
	// PushTx, PushBlock
	Status string `json:"status,omitempty"`
}

// Extra ifs for improved "security", right now it just checks version.
func SecurityCheck(req *Request) *Response {
	if req.Version == "" || req.Version != config.Get().Version {
		return &Response{Secure: false, Error: "version"}
	}
	return &Response{Secure: true, Error: "ok"}
}

func BlockCount(req *Request, db *types.DB) *Response {
	if db.Length >= 0 {
		return &Response{Length: db.Length, RecentHash: db.RecentHash, DiffLength: db.DiffLength}
	}
	return &Response{Length: -1, RecentHash: 0, DiffLength: "0"}
}

func RangeRequest(req *Request, db *types.DB) *Response {
	var (
		resp    Response
		counter int
	)

	for tools.JSONLen(resp) < config.Get().MaxDownload && req.Range[0]+counter <= req.Range[1] {
		block := db.GetBlock(req.Range[0] + counter)
		// if "length" in block: out.append(block)
		resp.Blocks = append(resp.Blocks, block)
		counter++
	}

	return &resp
}

func Txs(req *Request, db *types.DB) *Response {
	var resp Response
	resp.Txs = db.Txs
	return &resp
}

func PushTx(req *Request, db *types.DB) *Response {
	db.SuggestedTxs = append(db.SuggestedTxs, req.Tx)
	return &Response{Status: "success"}
}

func PushBlock(req *Request, db *types.DB) *Response {
	db.SuggestedBlocks = append(db.SuggestedBlocks, req.Block)
	return &Response{Status: "success"}
}
