package consensus

import (
	"log"
	"time"

	"github.com/toqueteos/altcoin/blockchain"
	"github.com/toqueteos/altcoin/config"
	"github.com/toqueteos/altcoin/server"
	"github.com/toqueteos/altcoin/tools"
	"github.com/toqueteos/altcoin/types"
)

func Run(db *types.DB, peers []string) {
	for _ = range time.Tick(1 * time.Second) {
		CheckPeers(db, peers)

		// Suggestions
		for _, tx := range db.SuggestedTxs {
			blockchain.AddTx(tx, db)
		}
		db.SuggestedTxs = nil

		for _, block := range db.SuggestedBlocks {
			blockchain.AddBlock(block, db)
		}
		db.SuggestedBlocks = nil
	}
}

// Check on the peers to see if they know about more blocks than we do.
func CheckPeers(db *types.DB, peers []string) {
	obj := &checkPeers{db, peers}

	for _, peer := range peers {
		// block_count := obj.cmd(peer, &server.Request{Type: "blockcount"})
		resp, err := server.SendCommand(peer, &server.Request{Type: "blockcount"})
		if err != nil {
			log.Println("[consensus.CheckPeers] blockcount request failed with error:", err)
			continue
		}

		// if not isinstance(block_count, dict): return
		// if "error" in block_count.keys(): return

		length := db.Length
		size := tools.Max(len(db.DiffLength), len(resp.DiffLength))
		us := tools.ZerosLeft(db.DiffLength, size)
		them := tools.ZerosLeft(resp.DiffLength, size)

		if them < us {
			obj.give_block(peer, resp.Length)
			continue
		}

		if us == them {
			obj.ask_for_txs(peer)
			continue
		}

		obj.download_blocks(peer, resp.Length, length)
	}
}

type checkPeers struct {
	db    *types.DB
	peers []string
}

func (obj *checkPeers) fork_check(newblocks []*types.Block) bool {
	block := obj.db.GetBlock(obj.db.Length)
	recent_hash := tools.DetHash(block)
	//their_hashes = map(tools.DetHash, newblocks)
	var their_hashes []string
	for _, b := range newblocks {
		their_hashes = append(their_hashes, tools.DetHash(b))
	}
	//return recent_hash not in their_hashes
	return tools.NotIn(recent_hash, their_hashes)
}

func (obj *checkPeers) bounds(length int, block_count int) []int {
	var end int
	if block_count-length > config.Get().DownloadMany {
		end = length + config.Get().DownloadMany - 1
	} else {
		end = block_count
	}
	return []int{tools.Max(length-2, 0), end}
}

func (obj *checkPeers) download_blocks(peer string, block_count int, length int) {
	resp, err := server.SendCommand(peer, &server.Request{Type: "range", Range: obj.bounds(length, block_count)})
	if err != nil || resp.Blocks == nil {
		log.Println("[consensus.download_blocks] range request failed with error:", err)
		return
	}

	// Only delete a max of 2 blocks, otherwise a peer might trick us into deleting everything over and over.
	for i := 0; i < 2; i++ {
		if obj.fork_check(resp.Blocks) {
			blockchain.DeleteBlock(obj.db)
		}
	}

	// DB['suggested_blocks'].extend(blocks)
	obj.db.SuggestedBlocks = append(obj.db.SuggestedBlocks, resp.Blocks...)
}

func (obj *checkPeers) ask_for_txs(peer string) {
	resp, err := server.SendCommand(peer, &server.Request{Type: "txs"})
	if err != nil {
		log.Println("[consensus.ask_for_txs] txs request failed with error:", err)
		return
	}

	// DB['suggested_txs'].extend(txs)
	obj.db.SuggestedTxs = append(obj.db.SuggestedTxs, resp.Txs...)

	// pushers = [x for x in DB['txs'] if x not in txs]
	// for push in pushers: cmd({'type': 'pushtx', 'tx': push})
	var pushers = make(map[*types.Tx]bool)
	for _, push := range obj.db.Txs {
		if _, ok := pushers[push]; !ok {
			if _, err := server.SendCommand(peer, &server.Request{Type: "pushtx", Tx: push}); err != nil {
				log.Println("[consensus.ask_for_txs] pushtx request failed with error:", err)
			}
			pushers[push] = true
		}
	}
}

func (obj *checkPeers) give_block(peer string, block_count int) {
	_, err := server.SendCommand(peer, &server.Request{Type: "pushblock", Block: obj.db.GetBlock(block_count + 1)})
	if err != nil {
		log.Println("[consensus.give_block] pushblock request failed with error:", err)
		return
	}
}
