package miner

import (
	"errors"
	"math/big"
	"math/rand"

	"github.com/toqueteos/altcoin/tools"
	"github.com/toqueteos/altcoin/types"
)

// Proof-of-Work
func PoW(block *types.Block, hashes int, restart chan bool) (bool, error) {
	hh := tools.DetHash(block)
	block.Nonce = randomNonce("100000000000000000")

	// count = 0
	var count int
	for tools.DetHash(&types.HalfWay{Nonce: block.Nonce, HalfHash: hh}) > block.Target {
		select {
		case <-restart:
			// return {"solution_found": true}
			return true, nil
		default:
			count++
			plus1(block.Nonce) // block.Nonce++

			if count > hashes {
				// return {"error": false}
				return false, errors.New("POW error")
			}

			// For testing sudden loss in hashpower from miners.
			// if block.Length > 150 {
			// } else {
			//     time.Sleep(10 * time.Millisecond) // 0.01 seconds
			// }
		}
	}

	return false, nil
}

var one = big.NewInt(1)

func plus1(n *big.Int) {
	n.Add(n, one)
}

func randomNonce(upto string) *big.Int {
	upperBound := new(big.Int)
	upperBound.SetString(upto, 10)

	nonce := new(big.Int)
	nonce.Rand(rand.New(rand.NewSource(99)), upperBound)

	return nonce
}
