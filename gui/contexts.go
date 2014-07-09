package gui

import (
	"github.com/toqueteos/altcoin/config"
)

var defaultCtx = Context{config.Get().CoinName}

type Context struct {
	CoinName string
}

type signErrorCtx struct {
	Context
	Err error
}

type amountErrorCtx struct {
	Context
	Amount string
}

type spendCtx struct {
	Context
	PrivKey      string
	Address      string
	CurrentBlock int
	Balance      float64
}
