package blockchain

import (
	"github.com/toqueteos/altcoin/tools"
	"github.com/toqueteos/altcoin/types"
)

// Returns the number of transactions that pubkey has broadcast.
func Count(addr string, db *types.DB) int {
	// def zeroth_confirmation_txs(address, DB):
	// 	def is_zero_conf(t):
	// 		return address == tools.make_address(t['pubkeys'], len(t['signatures']))
	// return len(filter(is_zero_conf, DB['txs']))
	var zerothConfirmationTxs int
	for _, t := range db.Txs {
		if addr == tools.MakeAddress(t.PubKeys, len(t.Signatures)) {
			zerothConfirmationTxs++
		}
	}

	current := db.GetAccount(addr)
	return current.Count + zerothConfirmationTxs
}
