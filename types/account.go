package types

import (
	"bytes"
	"encoding/json"
)

type Account struct {
	Amount int `json:"amount,omitempty"`
	Count  int `json:"count,omitempty"`
}

func (acc *Account) JSON() string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.Encode(acc)

	return buf.String()
}
