package miner

import (
	"time"
	"github.com/toqueteos/altcoin/types"
)

type Work struct {
	candidate_block   *types.Block
	hashes_till_check int
}

type Worker struct {
	Restart     chan bool
	SubmitQueue chan *types.Block
	WorkQueue   chan Work
}

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
