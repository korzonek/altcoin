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
	for _ = range time.Tick(config.Get().CheckPeersEvery) {
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
			obj.giveBlock(peer, resp.Length)
			continue
		}

		if us == them {
			obj.askForTxs(peer)
			continue
		}

		obj.downloadBlocks(peer, resp.Length, length)
	}
}

type checkPeers struct {
	db    *types.DB
	peers []string
}

func (obj *checkPeers) forkCheck(newblocks []*types.Block) bool {
	block := obj.db.GetBlock(obj.db.Length)
	//their_hashes = map(tools.DetHash, newblocks)
	recentHash := tools.DetHash(block)
	var theirHashes []string
	for _, b := range newblocks {
		theirHashes = append(theirHashes, tools.DetHash(b))
	}
	//return recent_hash not in their_hashes
	return tools.NotIn(recentHash, theirHashes)
}

func (obj *checkPeers) bounds(length int, blockCount int) []int {
	var end int
	if blockCount-length > config.Get().DownloadMany {
		end = length + config.Get().DownloadMany - 1
	} else {
		end = blockCount
	}
	return []int{tools.Max(length-2, 0), end}
}

func (obj *checkPeers) downloadBlocks(peer string, blockCount int, length int) {
	resp, err := server.SendCommand(peer, &server.Request{Type: "range", Range: obj.bounds(length, blockCount)})
	if err != nil || resp.Blocks == nil {
		log.Println("[consensus.downloadBlocks] range request failed with error:", err)
		return
	}

	// Only delete a max of 2 blocks, otherwise a peer might trick us into deleting everything over and over.
	for i := 0; i < 2; i++ {
		if obj.forkCheck(resp.Blocks) {
			blockchain.DeleteBlock(obj.db)
		}
	}

	// DB['suggested_blocks'].extend(blocks)
	obj.db.SuggestedBlocks = append(obj.db.SuggestedBlocks, resp.Blocks...)
}

func (obj *checkPeers) askForTxs(peer string) {
	resp, err := server.SendCommand(peer, &server.Request{Type: "txs"})
	if err != nil {
		log.Println("[consensus.askForTxs] txs request failed with error:", err)
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
				log.Println("[consensus.askForTxs] pushtx request failed with error:", err)
			}
			pushers[push] = true
		}
	}
}

func (obj *checkPeers) giveBlock(peer string, blockCount int) {
	_, err := server.SendCommand(peer, &server.Request{Type: "pushblock", Block: obj.db.GetBlock(blockCount + 1)})
	if err != nil {
		log.Println("[consensus.giveBlock] pushblock request failed with error:", err)
		return
	}
}
