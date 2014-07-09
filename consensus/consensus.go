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

func Run(peers []types.Peer, db *types.DB) {
	for _ = range time.Tick(1 * time.Second) {
		CheckPeers(peers, db)

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
func CheckPeers(peers []types.Peer, db *types.DB) {
	obj := &checkPeers{peers, db}

	for _, peer := range peers {
		block_count := obj.cmd(peer, &server.Request{Type: "blockcount"})

		// if not isinstance(block_count, dict): return
		// if "error" in block_count.keys(): return

		length := db.Length
		size := tools.Max(len(db.DiffLength), len(block_count.DiffLength))
		us := tools.ZerosLeft(db.DiffLength, size)
		them := tools.ZerosLeft(block_count.DiffLength, size)

		if them < us {
			obj.give_block(peer, block_count.Length)
			continue
		}

		if us == them {
			obj.ask_for_txs(peer)
			continue
		}

		obj.download_blocks(peer, block_count.Length, length)
	}
}

type checkPeers struct {
	peers []types.Peer
	db    *types.DB
}

func (obj *checkPeers) cmd(peer types.Peer, req *server.Request) *server.Response {
	resp, err := server.SendCommand(peer, req)
	if err != nil {
		log.Println(err)
		return nil
	}
	return resp
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

func (obj *checkPeers) download_blocks(peer types.Peer, block_count int, length int) {
	resp := obj.cmd(peer, &server.Request{Type: "range", Range: obj.bounds(length, block_count)})

	if resp.Blocks == nil {
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

func (obj *checkPeers) ask_for_txs(peer types.Peer) {
	resp := obj.cmd(peer, &server.Request{Type: "txs"})

	// DB['suggested_txs'].extend(txs)
	obj.db.SuggestedTxs = append(obj.db.SuggestedTxs, resp.Txs...)

	// pushers = [x for x in DB['txs'] if x not in txs]
	// for push in pushers: cmd({'type': 'pushtx', 'tx': push})
	var pushers = make(map[*types.Tx]bool)
	for _, push := range obj.db.Txs {
		if _, ok := pushers[push]; !ok {
			obj.cmd(peer, &server.Request{Type: "pushtx", Tx: push})
			pushers[push] = true
		}
	}

	//return []
}

func (obj *checkPeers) give_block(peer types.Peer, block_count int) {
	obj.cmd(peer, &server.Request{
		Type:  "pushblock",
		Block: obj.db.GetBlock(block_count + 1),
	})

	//return []
}
