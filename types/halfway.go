package types

import (
	"bytes"
	"encoding/json"
	"math/big"
)

type HalfWay struct {
	HalfHash string   `json:"halfhash,omitempty"`
	Nonce    *big.Int `json:"nonce,omitempty"`
}

func (h *HalfWay) Hash() string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.Encode(h)

	return buf.String()
}
