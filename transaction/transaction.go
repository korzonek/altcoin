// This file explains how we tell if a transaction is valid or not, it explains how we update the system when new transactions are added to the blockchain.

package transaction

import (
	"github.com/toqueteos/altcoin/config"
	"github.com/toqueteos/altcoin/tools"
	"github.com/toqueteos/altcoin/types"

	"github.com/conformal/btcec"
)

func SpendVerify(tx *types.Tx, txs []*types.Tx, db *types.DB) bool {
	txCopy := tx
	txCopy2 := tx

	// tx_copy.pop("signatures")
	txCopy.Signatures = nil

	if len(tx.PubKeys) == 0 {
		return false
	}

	if len(tx.Signatures) > len(tx.PubKeys) {
		return false
	}

	msg := tools.DetHash(txCopy)
	if !sigsMatch(tx.Signatures, tx.PubKeys, msg) {
		return false
	}

	if tx.Amount < config.Get().Fee {
		return false
	}

	address := addr(txCopy2)
	totalCost := 0

	//for Tx in filter(lambda t: address == addr(t), [tx] + txs) {
	for _, t := range append(txs, tx) {
		if address == addr(t) {
			continue
		}
		if t.Type == "spend" {
			totalCost += t.Amount
		}
		if t.Type == "mint" {
			totalCost -= config.Get().BlockReward
		}
	}

	return db.GetAccount(address).Amount >= totalCost
}

// def sigs_match(sigs, pubs, msg):
// 	return all(tools.verify(msg, sig, pub) for sig in sigs for pub in pubs)
func sigsMatch(sigs []*btcec.Signature, pubs []*btcec.PublicKey, msg string) bool {
	m := []byte(msg)
	for _, sig := range sigs {
		for _, pub := range pubs {
			if !tools.Verify(m, sig, pub) {
				return false
			}
		}
	}
	return true
}

func MintVerify(tx *types.Tx, txs []*types.Tx, db *types.DB) bool {
	//return 0 == len(filter(lambda t: t["type"] == "mint", txs))
	var n int
	for _, t := range txs {
		if t.Type == "mint" {
			n++
		}
	}
	return 0 == n
}

func Mint(tx *types.Tx, db *types.DB) {
	address := addr(tx)
	adjust("amount", address, config.Get().BlockReward, db)
	adjust("count", address, 1, db)
}

func Spend(tx *types.Tx, db *types.DB) {
	address := addr(tx)
	adjust("amount", address, -tx.Amount, db)
	adjust("amount", tx.To, tx.Amount-config.Get().Fee, db)
	adjust("count", address, 1, db)
}

func addr(tx *types.Tx) string {
	return tools.MakeAddress(tx.PubKeys, len(tx.Signatures))
}

// adjust(key, pubkey, amount, DB, sign=1)
func adjust(key string, addr string, value int, db *types.DB) {
	var sign = 1

	acc := db.GetAccount(addr)
	if !db.AddBlock {
		sign = -1
	}

	switch key {
	case "amount":
		acc.Amount += value * sign
	case "count":
		acc.Count += value * sign
	}

	db.Put(addr, acc)
}
