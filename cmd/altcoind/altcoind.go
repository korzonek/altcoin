package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/toqueteos/altcoin/config"
	"github.com/toqueteos/altcoin/consensus"
	"github.com/toqueteos/altcoin/gui"
	"github.com/toqueteos/altcoin/miner"
	"github.com/toqueteos/altcoin/server"
	"github.com/toqueteos/altcoin/tools"
	"github.com/toqueteos/altcoin/types"

	"github.com/syndtr/goleveldb/leveldb"
)

var (
	DatabaseFile     = "altcoin.db"
	WalletPassphrase = "my-altcoin-coin-wallet"
)

func main() {
	logger := log.New(os.Stdout, "[altcoind] ", log.Ldate|log.Ltime)

	// Create/Open a LevelDB database
	ldb, err := leveldb.OpenFile(DatabaseFile, nil)
	if err != nil {
		logger.Fatalf("Couldn't open %q\n", DatabaseFile)
	}

	// Create a *types.DB instance, this struct is passed around almost everywhere.
	// It holds a pointer to the LevelDB database among other things.
	db := types.NewDB(ldb)

	// List of peers we want to connect
	peers := []string{
		"localhost:8901",
		"localhost:8902",
		"localhost:8903",
		"localhost:8904",
		"localhost:8905",
	}

	//
	privkey := tools.DetHashString(WalletPassphrase)
	_, rewardAddress := tools.ParseKeyPair(privkey)

	// Let's setup ourselves as an altcoin node...
	cfg := config.DefaultConfig
	cfg.Version = "ALCv1.0"
	config.Set(cfg)

	// Setup done, now let's init the services and call it a day...
	go consensus.Run(db, peers)
	// Listens for peers. Peers might ask us for our blocks and our pool of recent transactions, or peers could suggest blocks and transactions to us.
	go server.Run(db)
	// Keeps track of blockchain database, checks on peers for new blocks and transactions.
	go miner.Run(db, peers, rewardAddress)
	// Browser based GUI.
	go gui.Run(db)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until a signal is received.
	<-c

	fmt.Println("Stopping altcoind...")
}
