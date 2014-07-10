package miner

import (
	"log"
	"runtime"
	"time"

	"github.com/toqueteos/altcoin/blockchain"
	"github.com/toqueteos/altcoin/config"
	"github.com/toqueteos/altcoin/tools"
	"github.com/toqueteos/altcoin/types"

	"github.com/conformal/btcec"
)

// Run spawns worker processes (multi-CPU mining) and coordinates the effort.
func Run(db *types.DB, peers []string, reward_address *btcec.PublicKey) {
	obj := &runner{
		db:             db,
		peers:          peers,
		reward_address: reward_address,
		submit_queue:   make(chan *types.Block),
	}

	// num_cores = multiprocessing.cpu_count()
	num_cores := runtime.NumCPU()
	log.Printf("Creating %d mining workers.", num_cores)
	for i := 0; i < num_cores; i++ {
		obj.workers = append(obj.workers, obj.spawn_worker())
	}

	var (
		candidate_block *types.Block
		length          int
	)

	for {
		length = db.Length
		if length == -1 {
			candidate_block = obj.genesis()
		} else {
			prev_block := db.GetBlock(length)
			candidate_block = obj.make_block(prev_block, db.Txs)
		}

		work := Work{candidate_block, config.Get().HashesPerCheck}
		for _, w := range obj.workers {
			w.WorkQueue <- work
		}

		// When block found, add to suggested blocks.
		// solved_block = submitted_blocks.get() # TODO(roasbeef): size=1?
		solved_block := <-obj.submit_queue
		if solved_block.Length != length+1 {
			continue
		}

		db.SuggestedBlocks = append(db.SuggestedBlocks, solved_block)
		obj.restart_workers()
	}
}

type runner struct {
	db             *types.DB
	peers          []string
	reward_address *btcec.PublicKey
	submit_queue   chan *types.Block
	workers        []*Worker
}

func (obj *runner) make_mint() *types.Tx {
	pubkeys := []*btcec.PublicKey{obj.reward_address}
	addr := tools.MakeAddress(pubkeys, 1)
	// TODO: `first_sig` should be a `config` var
	sign, err := btcec.ParseSignature([]byte("first_sig"), btcec.S256())
	if err != nil {
		log.Println("ParseSignature error:", err)
		return nil
	}

	return &types.Tx{
		Type:       "mint",
		PubKeys:    pubkeys,
		Signatures: []*btcec.Signature{sign},
		Count:      blockchain.Count(addr, obj.db),
	}
}

func (obj *runner) genesis() *types.Block {
	target := blockchain.Target(obj.db, 0)
	block := &types.Block{
		Version:    config.Get().Version,
		Length:     0,
		Time:       time.Now(), // time.time(),
		Target:     target,
		DiffLength: blockchain.HexInv(target),
		Txs:        []*types.Tx{obj.make_mint()},
	}
	log.Println("Genesis Block:", block)
	//block = tools.unpackage(tools.package(block))
	return block
}

func (obj *runner) make_block(prev_block *types.Block, txs []*types.Tx) *types.Block {
	length := prev_block.Length + 1
	target := blockchain.Target(obj.db, length)
	diffLength := blockchain.HexSum(prev_block.DiffLength, blockchain.HexInv(target))
	out := &types.Block{
		Version:    config.Get().Version,
		Txs:        append(txs, obj.make_mint()),
		Length:     length,
		Time:       time.Now(), // time.time(),
		DiffLength: diffLength,
		Target:     target,
		PrevHash:   tools.DetHash(prev_block),
	}
	//out = tools.unpackage(tools.package(out))
	return out
}

func (obj *runner) restart_workers() {
	log.Println("Possible solution found, restarting mining workers.")
	for _, w := range obj.workers {
		// worker_mailbox.["restart"].set()
		w.Restart <- true
	}
}

func (obj *runner) spawn_worker() *Worker {
	log.Println("Spawning worker")

	w := &Worker{
		Restart:     make(chan bool),
		SubmitQueue: obj.submit_queue,
		WorkQueue:   make(chan Work),
	}

	go Miner(w)

	return w
}
