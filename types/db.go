package types

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/syndtr/goleveldb/leveldb"
)

func NewDB(db *leveldb.DB) *DB {
	if db == nil {
		panic("Please provide a valid LevelDB database, not 'nil'.")
	}

	return &DB{
		Length:    -1,
		SigLength: -1,
		Storage:   db,
	}
}

type DB struct {
	AddBlock        bool
	DiffLength      string
	Length          int
	RecentHash      int
	SigLength       int
	Storage         *leveldb.DB
	SuggestedBlocks []*Block
	SuggestedTxs    []*Tx
	Txs             []*Tx
}

// def db_get(n, DB):
//     n = str(n)
//     try:
//         return tools.unpackage(DB['db'].Get(n))
//     except:
//         db_put(n, {'count': 0, 'amount': 0}, DB)  # Everyone defaults with
//         # having zero money, and having broadcast zero transcations.
//         return db_get(n, DB)

func (db *DB) GetBlock(blockNum int) *Block {
	num := strconv.Itoa(blockNum)
	key := []byte(num)

	value, err := db.Storage.Get(key, nil)
	if err != nil {
		return nil
	}

	var b Block
	if err := json.Unmarshal(value, &b); err != nil {
		return nil
	}

	return &b
}

func (db *DB) GetAccount(addr string) *Account {
	key := []byte(addr)

	value, err := db.Storage.Get(key, nil)
	switch err {
	case leveldb.ErrNotFound:
		// // Everyone defaults with having zero money, and having broadcast zero transactions.
		// db_put(n, {"count": 0, "amount": 0}, db)
		// return db_get(n, db)
		db.Put(addr, &Account{Count: 0, Amount: 0})
		return db.GetAccount(addr)
	case nil:
		// Nothing!
	default:
		log.Println("GetAccount error:", err)
		return nil
	}

	var acc Account
	if err := json.Unmarshal(value, &acc); err != nil {
		log.Println("json.Unmarshal error:", err)
		return nil
	}
	return &acc
}

// def db_put(key, dic, DB):
//     return DB['db'].Put(str(key), tools.package(dic))

func (db *DB) Put(k string, v Serializer) error {
	key := []byte(k)
	value := []byte(v.JSON())
	return db.Storage.Put(key, value, nil)
}

// def db_delete(key, DB):
//     return DB['db'].Delete(str(key))

func (db *DB) Delete(k string) error {
	key := []byte(k)
	return db.Storage.Delete(key, nil)
}
