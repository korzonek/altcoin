package types

import (
	"bytes"
	"encoding/json"
	"math/big"
	"time"
)

type Block struct {
	DiffLength string    `json:"difflength,omitempty"`
	Error      error     `json:"error,omitempty"`
	Length     int       `json:"length,omitempty"`
	Nonce      *big.Int  `json:"nonce,omitempty"`
	PrevHash   string    `json:"prevhash,omitempty"`
	Target     string    `json:"target,omitempty"`
	Time       time.Time `json:"time,omitempty"`
	Txs        []*Tx     `json:"txs,omitempty"`
	Version    string    `json:"version,omitempty"`
}

func (b *Block) Hash() string {
	return b.JSON()
}

func (b *Block) JSON() string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.Encode(b)

	return buf.String()
}
