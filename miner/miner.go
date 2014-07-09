package miner

import (
	"errors"
	"math/big"
	"math/rand"
	"time"

	"github.com/toqueteos/altcoin/tools"
	"github.com/toqueteos/altcoin/types"
)

func Miner(worker *Worker) {
	var (
		block             *types.Block
		hashes_till_check int
		// need_new_work = false
		need_new_work bool
	)

	for {
		// # Either get the current block header, or restart because a block has
		// # been solved by another worker.
		// try:
		//     if need_new_work or block is None:
		//         block, hashes_till_check = workQueue.get(True, 1)
		//         need_new_work = False
		// # Try to optimistically get the most up-to-date work.
		// except Empty:
		//     need_new_work = False
		//     continue
		if need_new_work || block == nil {
			select {
			case work := <-worker.WorkQueue:
				block, hashes_till_check = work.candidate_block, work.hashes_till_check
				need_new_work = false
			case <-time.After(1 * time.Second):
				need_new_work = false
				continue
			}
		}

		solution_found, err := PoW(block, hashes_till_check, worker.Restart)

		switch {
		// We hit the hash ceiling.
		case err != nil:
		// Another worker found the block.
		case solution_found:
			// Empty out the signal queue.
			need_new_work = true
		// Block found!
		default:
			worker.SubmitQueue <- block
			need_new_work = true
		}
	}
}

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
