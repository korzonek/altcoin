// This file explains how we talk to the database. It explains the rules for adding blocks and transactions.

package blockchain

import (
	"log"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/toqueteos/altcoin/config"
	"github.com/toqueteos/altcoin/tools"
	"github.com/toqueteos/altcoin/transaction"
	"github.com/toqueteos/altcoin/types"
)

var (
	transaction_keys = []string{"mint", "spend"}

	transaction_update = map[string]func(*types.Tx, *types.DB){
		"mint":  transaction.Mint,
		"spend": transaction.Spend,
	}

	transaction_verify = map[string]func(*types.Tx, []*types.Tx, *types.DB) bool{
		"mint":  transaction.MintVerify,
		"spend": transaction.SpendVerify,
	}

	targets = map[int]string{}
	times   = map[int]float64{}
)

// RecentBlockTargets grabs info from old blocks.
// recent_blockthings(key, DB, size, length=0)
func RecentBlockTargets(db *types.DB, size, length int) []string {
	if length == 0 {
		length = db.Length
	}

	start := length - size
	if start < 0 {
		start = 0
	}

	var ts []string
	for index := start; index < length; index++ {
		// if not index in storage:
		//     storage[index] = db_get(index, db)["target"]
		_, ok := targets[index]
		if !ok {
			targets[index] = db.GetBlock(index).Target
		}

		ts = append(ts, targets[index])
	}

	return ts
}

func RecentBlockTimes(db *types.DB, size, length int) []float64 {
	if length == 0 {
		length = db.Length
	}

	start := length - size
	if start < 0 {
		start = 0
	}

	var ts []float64
	for index := start; index < length; index++ {
		// if not index in storage:
		//     storage[index] = db_get(index, db)["target"]
		_, ok := times[index]
		if !ok {
			t := db.GetBlock(index).Time
			times[index] = unix(t)
		}

		ts = append(ts, times[index])
	}

	return ts
}

// Returns the target difficulty at a paticular blocklength.
// target(DB, length=0)
func Target(db *types.DB, length int) string {
	if length == 0 {
		length = db.Length
	}

	// Use same difficulty for first few blocks.
	if length < 4 {
		return strings.Repeat("0", 4) + strings.Repeat("f", 60)
	}

	if length <= db.Length {
		return targets[length] // Memoized
	}

	weights := func(length int) []float64 {
		var r []float64
		for i := 0; i < length; i++ {
			exp := float64(length - i)
			r = append(r, math.Pow(config.Get().Inflection, exp))
		}
		return r
	}

	// We are actually interested in the average number of hashes required to mine a block.
	// Number of hashes required is inversely proportional to target.
	// So we average over inverse-targets, and inverse the final answer.
	estimate_target := func(db *types.DB) string {
		sumTargets := func(ls []string) string {
			if len(ls) < 1 {
				return "0" // must be string, int on python version
			}

			// This is basically an ofuscated REDUCE
			// while len(ls) > 1:
			// 	ls = [hexSum(ls[0], ls[1])] + ls[2:]
			// return ls[0]
			var r = HexSum(ls[0], ls[1])
			for _, elem := range ls[2:] {
				r = HexSum(r, elem)
			}
			return r
		}

		blocks := RecentBlockTargets(db, config.Get().HistoryLength, 0)
		w := weights(len(blocks))
		//tw = sum(w)
		var tw float64
		for _, welem := range w {
			tw += welem
		}

		var targets []string
		for _, t := range blocks {
			targets = append(targets, HexInv(t))
		}

		weighted_multiply := func(i int) string {
			return HexMul(targets[i], strconv.Itoa(int(w[i]/tw)))
		}

		var weighted_targets []string
		for i := 0; i < len(targets); i++ {
			weighted_targets = append(weighted_targets, weighted_multiply(i))
		}

		return HexInv(sumTargets(weighted_targets))
	}

	estimate_time := func(db *types.DB) float64 {
		times := RecentBlockTimes(db, config.Get().HistoryLength, 0)

		var lengths []float64
		for i := 1; i < len(times); i++ {
			lengths = append(lengths, times[i]-times[i-1])
		}

		// Geometric weighting
		w := weights(len(lengths))

		// Normalization constant
		// tw = sum(w)
		var tw float64
		for _, elem := range w {
			tw += elem
		}

		var r []float64
		for i := 0; i < len(lengths); i++ {
			r = append(r, w[i]*lengths[i]/tw)
		}

		// sum(r)
		var sum float64
		for _, elem := range r {
			sum += elem
		}

		return sum
	}

	block_time := config.BlockTime(length)
	retarget := estimate_time(db) / float64(block_time)
	return HexMul(estimate_target(db), strconv.Itoa(int(retarget)))
}

// Attempts adding a new block to the blockchain.
func AddBlock(block *types.Block, db *types.DB) {
	tx_check := func(txs []*types.Tx) bool {
		// start = copy.deepcopy(txs)
		var start = txs
		var txs_source []*types.Tx
		var start_copy []*types.Tx

		for !reflect.DeepEqual(start, start_copy) {
			// Block passes this test
			if start == nil {
				return false
			}

			// start_copy = copy.deepcopy(start)
			start_copy = start
			last := start[len(start)-1]

			// transactions.tx_check[start[-1]['type']](start[-1], out, DB)
			fn := transaction_verify[last.Type]
			if fn(last, txs_source, db) {
				// start.pop()
				start = start[:len(start)-1]
				txs_source = append(txs_source, last)
			} else {
				// Block is invalid
				return true
			}
		}

		// Block is invalid
		return true
	}

	// if "error" in block: return False
	if block.Error != nil {
		return
	}

	// if "length" not in block: return False
	// NOTE: block.Length not being set means it takes its "zero value".
	// This shouldn't be a problem, check out next if stmt.
	if block.Length == 0 {
		return
	}

	length := db.Length
	if block.Length != length+1 {
		return
	}

	if block.DiffLength != HexSum(db.DiffLength, HexInv(block.Target)) {
		return
	}

	if length >= 0 && tools.DetHash(db.GetBlock(length)) != block.PrevHash {
		return
	}

	// a = copy.deepcopy(block)
	// a.pop("nonce")
	block_copy := block
	block_copy.Nonce = nil

	//if "target" not in block.keys(): return False
	if block.Target == "" {
		return
	}

	half_way := &types.HalfWay{
		Nonce:    block.Nonce,
		HalfHash: tools.DetHash(block_copy),
	}

	if tools.DetHash(half_way) > block.Target {
		return
	}

	if block.Target != Target(db, block.Length) {
		return
	}

	// TODO: Figure out why 8 (length)?
	earliest_median := median(RecentBlockTimes(db, config.Get().Mmm, 8))
	// `float64` (unix epoch) back to `time.Time`
	sec, nsec := math.Modf(earliest_median)
	earliest := time.Unix(int64(sec), int64(nsec*1e9))

	// if block.Time > time.time(): return false
	// if block.Time < earliest: return false
	if block.Time.After(time.Now()) || block.Time.Before(earliest) {
		return
	}

	if tx_check(block.Txs) {
		return
	}

	// block_check was unnecessary because it was only called once
	// and it only returned true at its end

	// if block_check(block, db):
	log.Println("add_block:", block)
	db.Put(strconv.Itoa(block.Length), block)

	db.Length = block.Length
	db.DiffLength = block.DiffLength

	orphans := db.Txs
	db.Txs = nil

	for _, tx := range block.Txs {
		db.AddBlock = true
		fn := transaction_update[tx.Type]
		fn(tx, db)
	}

	for _, tx := range orphans {
		AddTx(tx, db)
	}
}

// DeleteBlock removes the most recent block from the blockchain.
func DeleteBlock(db *types.DB) {
	if db.Length < 0 {
		return
	}

	// try:
	// 	targets.pop(str(DB['length']))
	// except:
	// 	pass
	// try:
	// 	times.pop(str(DB['length']))
	// except:
	// 	pass
	delete(targets, db.Length)
	delete(times, db.Length)

	block := db.GetBlock(db.Length)
	orphans := sortedOrphans(db.Txs)
	db.Txs = nil

	for _, tx := range block.Txs {
		orphans = append(orphans, tx)
		db.AddBlock = false
		fn := transaction_update[tx.Type]
		fn(tx, db)
	}

	db.Delete(strconv.Itoa(db.Length))
	db.Length -= 1

	if db.Length == -1 {
		db.DiffLength = "0"
	} else {
		block = db.GetBlock(db.Length)
		db.DiffLength = block.DiffLength
	}

	// for orphan in sorted(orphans, key=lambda x: x["count"]):
	sort.Sort(orphans)
	for _, orphan := range orphans {
		AddTx(orphan, db)
	}
}
