package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/toqueteos/altcoin/config"
	"github.com/toqueteos/altcoin/types"

	"github.com/btcsuite/btcutil/base58"
	"github.com/conformal/btcec"
	"github.com/conformal/btcwire"
)

func Sign(msg []byte, privkey *btcec.PrivateKey) (*btcec.Signature, error) {
	h := btcwire.DoubleSha256(msg)
	return privkey.Sign(h)
}

func Verify(msg []byte, sig *btcec.Signature, pubkey *btcec.PublicKey) bool {
	h := btcwire.DoubleSha256(msg)
	return sig.Verify(h, pubkey)
}

func ParseKeyPair(privkey string) (*btcec.PrivateKey, *btcec.PublicKey) {
	return btcec.PrivKeyFromBytes(btcec.S256(), []byte(privkey))
}

// Deterministically takes hash (defaults to sha256) of dict, list, int, or string.
// String representations, (what is feed into hash fn, in python+go pseudocode):
// - list: fmt.Sprintf("[%s]", ",".join(map(det, sorted(payload))))
// - dict: fmt.Sprintf("{%s}", ",".join(map(lambda p: det(p[0]) + ":" + det(p[1]), sorted(x.items()))))
// - int: str(payload)
// - string: payload
// return hash_(json.loads(json.dumps(payload_string))))
func DetHash(h types.Hasher) string { return config.Hash(h.Hash()) }
func DetHashInt(h int) string       { return config.Hash(strconv.Itoa(h)) }
func DetHashString(h string) string { return config.Hash(h) }

// n is the number of pubkeys required to spend from this address.
func MakeAddress(pubkeys []*btcec.PublicKey, n int) string {
	addr := &types.Address{N: n, PubKeys: pubkeys}
	h := DetHash(addr)
	b58 := base58.Encode([]byte(h))
	return fmt.Sprintf("%d%d%x", len(pubkeys), n, b58[:29])
}

func ZerosLeft(s string, size int) string {
	qty := size - len(s)
	if qty > 0 {
		zeros := strings.Repeat("0", qty)
		return zeros + s
	}
	return s
}

func In(s string, cases []string) bool {
	for _, elem := range cases {
		if elem == s {
			return true
		}
	}
	return false
}

func NotIn(s string, ss []string) bool {
	return !In(s, ss)
}

func JSONLen(value interface{}) int {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	err := enc.Encode(value)
	if err != nil {
		return -1
	}

	// \n is always added to the end, substract that
	s := buf.String()
	return len(s) - 1
}

func Max(a, b int) int {
	if a < b {
		return b
	}
	return a
}
