// The easiest way to understand this file is to try it out and have a look at the html it creates.
// It creates a very simple page that allows you to spend money.

package gui

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/toqueteos/altcoin/blockchain"
	"github.com/toqueteos/altcoin/config"
	"github.com/toqueteos/altcoin/tools"
	"github.com/toqueteos/altcoin/types"

	"github.com/conformal/btcec"

	"github.com/codegangsta/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
)

func GetHome(ren render.Render) {
	ren.HTML(200, "wallet", defaultCtx)
}

func PostHome(req *http.Request, ren render.Render) {
	// Hashed form input is used as private key and stored on session.
	wallet := req.FormValue("wallet")
	h := tools.DetHashString(wallet)
	// Redirec to wallet spend page
	ren.Redirect("/spend/"+h, http.StatusOK)
}

// RequireWallet only returns a template if privkey isn't an URL param. See *Spend fns.
func RequireWallet(params martini.Params, ren render.Render) {
	if params["privkey"] != "" {
		return
	}

	ren.HTML(200, "error", defaultCtx)
}

func GetSpend(db *types.DB, params martini.Params, ren render.Render) {
	privkey := params["privkey"]
	// Get our own public key
	_, pubkey := tools.ParseKeyPair(privkey)

	// TODO: Calculate our own address every time?
	addr := tools.MakeAddress([]*btcec.PublicKey{pubkey}, 1)
	// TODO: Some sort of balance cache would be nice
	// (instead of traversing the entire blockchain).
	balance := db.GetAccount(addr).Amount
	for _, tx := range db.Txs {
		if tx.Type == "spend" && tx.To == addr {
			balance += tx.Amount - config.Get().Fee
		}
		if tx.Type == "spend" && tx.PubKeys[0] == pubkey {
			balance -= tx.Amount
		}
	}

	ren.HTML(200, "spend", spendCtx{
		Context:      defaultCtx,
		PrivKey:      privkey,
		Address:      addr,
		CurrentBlock: db.Length,
		Balance:      float64(balance) / 100000.0,
	})
}

// /spend/:privkey
func PostSpend(db *types.DB, params martini.Params, req *http.Request, ren render.Render) {
	privkey := params["privkey"]
	// Form input
	formAmount := req.FormValue("amount")
	formTo := req.FormValue("to")

	amount, err := strconv.Atoi(formAmount)
	if err != nil {
		ren.HTML(200, "errors/amount", amountErrorCtx{defaultCtx, formAmount})
		return
	}

	if err := spend(db, amount, privkey, formTo); err != nil {
		ren.HTML(200, "errors/sign", signErrorCtx{defaultCtx, err})
	}

	ren.Redirect("/spend/"+privkey, http.StatusOK)
}

func Run(db *types.DB) {
	m := martini.New()
	m.Use(martini.Logger())
	m.Use(martini.Recovery())
	m.Use(martini.Static("public"))

	r := martini.NewRouter()
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)

	m.Use(render.Renderer(render.Options{
		Layout:     "layout",
		Extensions: []string{".tmpl", ".html"},
	}))
	m.Map(db)

	store := sessions.NewCookieStore(config.Get().GuiSessionKeyPairs...)
	m.Use(sessions.Sessions(config.Get().CoinName+"_session", store))

	r.Get("/", GetHome)
	r.Get("/home", GetHome)
	r.Post("/home", PostHome)

	r.Get("/spend/:privkey", RequireWallet, GetSpend)
	r.Post("/spend/:privkey", RequireWallet, PostSpend)

	if !config.Get().UseSSL {
		// HTTP
		var httpAddr = fmt.Sprintf(":%d", config.Get().GuiPort)
		if err := http.ListenAndServe(httpAddr, m); err != nil {
			log.Fatal(err)
		}
	} else {
		// HTTPS
		// To generate a development cert and key, run the following from your *nix terminal:
		// go run $GOROOT/src/pkg/crypto/tls/generate_cert.go --host="localhost"
		var httpsAddr = fmt.Sprintf(":%d", config.Get().GuiPortSSL)
		if err := http.ListenAndServeTLS(httpsAddr, "cert.pem", "key.pem", m); err != nil {
			log.Fatal(err)
		}
	}
}

// spend adds a tx which represents `from` paying `amount` coins to `to`.
// Both `from` and `to` are the string version of PrivateKey and PublicKey
// of the sender and receiver, respectively.
func spend(db *types.DB, amount int, from string, to string) error {
	amount = amount * 100000 // or: amount *= 100000

	privkey, pubkey := tools.ParseKeyPair(from)
	pubkeys := []*btcec.PublicKey{pubkey}
	addr := tools.MakeAddress(pubkeys, 1)

	tx := &types.Tx{
		Type:    "spend",
		PubKeys: pubkeys,
		Amount:  amount,
		To:      to,
	}

	// try:
	//     tx["count"] = blockchain.count(address, db)
	// except:
	//     tx["count"] = 1
	// Why try .. except?
	tx.Count = blockchain.Count(addr, db)

	htx := []byte(tools.DetHash(tx))
	sign, err := privkey.Sign(htx)
	if err != nil {
		return err
	}

	tx.Signatures = []*btcec.Signature{sign}
	log.Println("Created Tx:", tx)
	blockchain.AddTx(tx, db)
	return nil
}
