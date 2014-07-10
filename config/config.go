package config

import (
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/conformal/btcec"
)

func Get() *Config { return currentConfig }
func Set(c *Config) {
	currentConfig = c
	if c.MintSignature == nil {
		log.Fatalln("Signature for mint transactions is required")
	}
}

var currentConfig *Config = DefaultConfig

type Config struct {
	CoinName string
	Version  string

	ListenPort int

	HashesPerCheck int
	BlockReward    int
	Premine        int
	Fee            int
	MintSignature  *btcec.Signature

	Mmm        int     // Lower limits on what the "time" tag in a block can say.
	Inflection float64 // This constant is selected such that the 50 most recent blocks count for 1/2 the total weight.

	DownloadMany int // Max number of blocks to request from a peer at the same time.
	MaxDownload  int
	// Take the median of this many blocks.
	// How far back in history do we look when we use statistics to guess at the
	// current blocktime and difficulty.
	HistoryLength int

	// Brainwallet string // "brain wallet"
	// Privatekey  string // Hash(Brainwallet)
	// Publickey   *btcec.PublicKey // _, pub := tools.ParseKeyPair(privkey)

	UseSSL             bool
	GuiPort            int
	GuiPortSSL         int
	GuiSessionKeyPairs [][]byte
}

var DefaultConfig = &Config{
	CoinName:       "AltCoin",
	Version:        "VERSION",
	ListenPort:     8900,
	HashesPerCheck: 100000,
	MintSignature:  nil,
	BlockReward:    100000,
	Premine:        5000000,
	Fee:            1000,
	Mmm:            100,
	Inflection:     0.985,
	DownloadMany:   500,
	MaxDownload:    50000,
	HistoryLength:  400,
	UseSSL:         false,
	GuiPort:        8080,
	GuiPortSSL:     8443,
	GuiSessionKeyPairs: [][]byte{
		[]byte("type-in-a-random-string-here"),
		[]byte("type-in-another-random-string-here"),
	},
}

var Hash = hash
var BlockTime = blockTime

// Hash takes sha256 hash of: (dict, list, int or str) supplied as a string.
func hash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func blockTime(length int) int {
	if length*Get().BlockReward < Get().Premine {
		return 30 // seconds
	} else {
		return 60
	}
}
