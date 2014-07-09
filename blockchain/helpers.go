package blockchain

import (
	"encoding/hex"
	"log"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/toqueteos/altcoin/tools"
	"github.com/toqueteos/altcoin/types"
)

// median is good for weeding out liars, so long as the liars don't have 51% hashpower.
func median(ts []float64) float64 {
	if len(ts) < 1 {
		return 0
	}

	sort.Float64s(ts)
	return ts[len(ts)/2]
}

// unix returns Unix's epoch time in Python format.
// Go's Unix (and UnixNano) returns an int64,
func unix(t time.Time) float64 {
	ns := float64(t.Nanosecond()) / 1e9
	return float64(t.Unix()) + ns
}

func hexBig(op, left, right string) string {
	na := new(big.Int)
	nb := new(big.Int)

	if _, ok := na.SetString(left, 16); !ok {
		log.Fatalln("invalid SetString input %s (left)", left)
	}
	if _, ok := nb.SetString(right, 16); !ok {
		log.Fatalln("invalid SetString input %s (right)", right)
	}

	switch op {
	// Sum of numbers expressed as hexidecimal strings
	case "sum":
		na.Add(na, nb)
	// Use double-size for division, to reduce information leakage.
	case "invert":
		na.Div(na, nb)
	case "times":
		na.Mul(na, nb)
	}
	num := hex.EncodeToString(na.Bytes())

	return tools.ZerosLeft(num, 64)
}

var hexInvertLeft = strings.Repeat("f", 128)

func HexSum(left, right string) string { return hexBig("sum", left, right) }
func HexInv(right string) string       { return hexBig("invert", hexInvertLeft, right) }
func HexMul(left, right string) string { return hexBig("times", left, right) }

type sortedOrphans []*types.Tx

func (o sortedOrphans) Len() int           { return len(o) }
func (o sortedOrphans) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o sortedOrphans) Less(i, j int) bool { return o[i].Count < o[j].Count }
