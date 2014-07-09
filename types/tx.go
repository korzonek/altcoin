package types

import (
	"bytes"
	"encoding/json"

	"github.com/conformal/btcec"
)

// Tx holds all related info for a transaction
// NOTE: Types int/int64 are probably not enough for Amount.
// Python has Long (big) numbers support builtin, Go doesn't, so "the big.Int way"
// Extracted from `gui.py:13`
type Tx struct {
	Amount     int                `json:"amount,omitempty"`
	Count      int                `json:"count,omitempty"`
	PubKeys    []*btcec.PublicKey `json:"pubkeys,omitempty"`
	Signatures []*btcec.Signature `json:"signatures,omitempty"`
	To         string             `json:"to,omitempty"`
	Type       string             `json:"type,omitempty"`
}

func (t *Tx) Hash() string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.Encode(t)

	return b.String()
}
