package miner

import (
	"log"
	"os"
	"runtime"
	"time"

	"github.com/toqueteos/altcoin/blockchain"
	"github.com/toqueteos/altcoin/config"
	"github.com/toqueteos/altcoin/tools"
	"github.com/toqueteos/altcoin/types"

	"github.com/conformal/btcec"
)

var logger = log.New(os.Stdout, "[miner] ", log.Ldate|log.Ltime|log.Lshortfile)

// Run spawns worker processes (multi-CPU mining) and coordinates the effort.
func Run(db *types.DB, peers []string, rewardAddr *btcec.PublicKey) {
	obj := &runner{
		db:         db,
		peers:      peers,
		rewardAddr: rewardAddr,
		submitCh:   make(chan *types.Block),
	}

	cpus := runtime.NumCPU()
	logger.Printf("Creating %d mining workers...", cpus)
	for i := 0; i < cpus; i++ {
		obj.workers = append(obj.workers, NewWorker(obj.submitCh))
		logger.Printf("Spawning worker %d...", i)
	}

	var (
		block  *types.Block
		length int
	)

	for {
		length = db.Length
		if length == -1 {
			block = obj.genesis()
		} else {
			prevBlock := db.GetBlock(length)
			block = obj.makeBlock(prevBlock, db.Txs)
		}

		work := Work{
			block:          block,
			hashesPerCheck: config.Get().HashesPerCheck,
		}
		for _, w := range obj.workers {
			w.WorkQueue <- work
		}

		// When block found, add to suggested blocks.
		solvedBlock := <-obj.submitCh
		if solvedBlock.Length != length+1 {
			continue
		}

		db.SuggestedBlocks = append(db.SuggestedBlocks, solvedBlock)

		// Restart workers
		logger.Println("Possible solution found, restarting mining workers.")
		for _, w := range obj.workers {
			w.Restart <- true
		}
	}
}

type runner struct {
	db         *types.DB
	peers      []string
	rewardAddr *btcec.PublicKey
	submitCh   chan *types.Block
	workers    []*Worker
}

func (obj *runner) makeMint() *types.Tx {
	pubkeys := []*btcec.PublicKey{obj.rewardAddr}
	addr := tools.MakeAddress(pubkeys, 1)

	return &types.Tx{
		Type:       "mint",
		PubKeys:    pubkeys,
		Signatures: []*btcec.Signature{nil},
		Count:      blockchain.Count(addr, obj.db),
	}
}

func (obj *runner) genesis() *types.Block {
	target := blockchain.Target(obj.db, 0)
	block := &types.Block{
		Version:    config.Get().Version,
		Length:     0,
		Time:       time.Now(),
		Target:     target,
		DiffLength: blockchain.HexInv(target),
		Txs:        []*types.Tx{obj.makeMint()},
	}
	logger.Println("Genesis Block:", block)
	return block
}

func (obj *runner) makeBlock(prevBlock *types.Block, txs []*types.Tx) *types.Block {
	length := prevBlock.Length + 1
	target := blockchain.Target(obj.db, length)
	diffLength := blockchain.HexSum(prevBlock.DiffLength, blockchain.HexInv(target))
	out := &types.Block{
		Version:    config.Get().Version,
		Txs:        append(txs, obj.makeMint()),
		Length:     length,
		Time:       time.Now(),
		DiffLength: diffLength,
		Target:     target,
		PrevHash:   tools.DetHash(prevBlock),
	}
	return out
}
