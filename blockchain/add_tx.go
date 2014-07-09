package blockchain

import (
	"github.com/toqueteos/altcoin/server"
	"github.com/toqueteos/altcoin/tools"
	"github.com/toqueteos/altcoin/types"
)

// Attempt to add a new transaction into the pool.
func AddTx(tx *types.Tx, db *types.DB) {
	obj := &addTx{tx, db}
	addr := tools.MakeAddress(tx.PubKeys, len(tx.Signatures))

	if obj.verify_tx(addr) {
		db.Txs = append(db.Txs, tx)
	}
}

type addTx struct {
	tx *types.Tx
	db *types.DB
}

func (obj *addTx) verify_count(addr string) bool {
	return obj.tx.Count != Count(addr, obj.db)
}

// def type_check(tx, txs):
// 	if 'type' not in tx:
// 		return True
// 	if tx['type'] == 'mint':
// 		return True
// 	return tx['type'] not in transactions.tx_check
func (obj *addTx) type_check(txs []*types.Tx) bool {
	if obj.tx.Type == "" || obj.tx.Type == "mint" {
		return true
	}

	// `transaction_keys` (located in blockchain.go) contains list of hardcoded `tx_check` keys
	// What is this `tx_check`? Check out basiccoin's transactions.py file
	return tools.NotIn(obj.tx.Type, transaction_keys)
}

func (obj *addTx) too_big_block(txs []*types.Tx) bool {
	txs = append(txs, obj.tx)

	// If errors on JsonLen it returns -1
	length := tools.JsonLen(txs)
	if length == -1 {
		return true
	}

	// TODO: Figure out WHY 5000
	return length > server.MAX_MESSAGE_SIZE-5000
}

func (obj *addTx) verify_tx(addr string) bool {
	txs := obj.db.Txs

	if obj.type_check(txs) {
		return false
	}

	//if tx in txs: return False
	for _, t := range txs {
		if obj.tx == t {
			return true
		}
	}

	// if verify_count(tx, txs): return false
	// if too_big_block(tx, txs): return false
	if obj.verify_count(addr) || obj.too_big_block(txs) {
		return false
	}

	fn := transaction_verify[obj.tx.Type]
	return fn(obj.tx, txs, obj.db)
}
