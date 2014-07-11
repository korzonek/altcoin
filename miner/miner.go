package miner

import (
	"time"

	"github.com/toqueteos/altcoin/types"
)

type Work struct {
	block          *types.Block
	hashesPerCheck int
}

type Worker struct {
	Restart     chan bool
	SubmitQueue chan *types.Block
	WorkQueue   chan Work
}

func NewWorker(submit chan *types.Block) *Worker {
	w := &Worker{
		Restart:     make(chan bool),
		SubmitQueue: submit,
		WorkQueue:   make(chan Work),
	}

	go Miner(w)

	return w
}

func Miner(worker *Worker) {
	var (
		block          *types.Block
		hashesPerCheck int
		// need_new_work = false
		needNewWork bool
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
		if needNewWork || block == nil {
			select {
			case work := <-worker.WorkQueue:
				block, hashesPerCheck = work.block, work.hashesPerCheck
				needNewWork = false
			case <-time.After(1 * time.Second):
				needNewWork = false
				continue
			}
		}

		solutionFound, err := PoW(block, hashesPerCheck, worker.Restart)

		switch {
		// We hit the hash ceiling.
		case err != nil:
		// Another worker found the block.
		case solutionFound:
			// Empty out the signal queue.
			needNewWork = true
		// Block found!
		default:
			worker.SubmitQueue <- block
			needNewWork = true
		}
	}
}
