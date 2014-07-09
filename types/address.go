package types

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/conformal/btcec"
)

type Address struct {
	N       int                `json:"n,omitempty"`
	PubKeys []*btcec.PublicKey `json:"pubkeys,omitempty"`
}

func (addr *Address) Hash() string {
	return fmt.Sprintf("%d:[%s]", addr.N, addr.Sorted())
}

func (addr *Address) Sorted() string {
	var pks []string
	for _, pub := range addr.PubKeys {
		s := hex.EncodeToString(pub.SerializeCompressed())
		pks = append(pks, s)
	}

	sort.Strings(pks)
	return strings.Join(pks, ",")
}
