package main

import (
	"crypto/sha512"
	"fmt"
	"log"

	"github.com/toqueteos/altcoin/config"
	"github.com/toqueteos/altcoin/consensus"
	"github.com/toqueteos/altcoin/gui"
	"github.com/toqueteos/altcoin/server"
	"github.com/toqueteos/altcoin/types"

	"github.com/syndtr/goleveldb/leveldb"
)

var DatabaseFile = "AltCoin.db"

func main() {
	// Create/Open a LevelDB database
	ldb, err := leveldb.OpenFile(DatabaseFile, nil)
	if err != nil {
		log.Fatalf("Couldn't open %q\n", DatabaseFile)
	}

	// Create a *types.DB instance, this struct is passed around almost everywhere.
	// It holds a pointer to the LevelDB database among other things.
	db := types.NewDB(ldb)

	peers := []types.Peer{
		types.Peer{"localhost", 8901},
		types.Peer{"localhost", 8902},
		types.Peer{"localhost", 8903},
		types.Peer{"localhost", 8904},
		types.Peer{"localhost", 8905},
	}

	// Let's say we want to change coin name and Hash function.
	cfg := config.DefaultConfig
	cfg.CoinName = "AwesomeCoin"
	// Let global config know about our new Hash function
	config.Hash = Sha512Hash

	go consensus.Run(peers, db)
	// Listens for peers. Peers might ask us for our blocks and our pool of recent transactions, or peers could suggest blocks and transactions to us.
	server.Run(db)
	// Keeps track of blockchain database, checks on peers for new blocks and transactions.
	//miner.Run(db)
	// Browser based GUI.
	gui.Run(db)
}

// This is our new Hash function (uses sha512 instead of sha256)
func Sha512Hash(s string) string {
	h := sha512.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
