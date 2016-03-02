# altcoin

[![GoDoc](https://godoc.org/github.com/toqueteos/altcoin?status.png)]
(http://godoc.org/github.com/toqueteos/altcoin)[![codebeat](https://codebeat.co/badges/a7b74900-1bbc-439b-a523-24215742994d)](https://codebeat.co/projects/github-com-toqueteos-altcoin)

The simplest crypto-currency ported from Python to Go for readability and type-safety, also for the fun and learn new things.

**DISCLAIMER:** Identifier names (variables and funcs) don't *strictly* follow Go styleguide, for example, you may find `var foo_bar` instead of `var fooBar`.
This would be fixed gradually.

## Installation

Tested on: **go version go1.3 windows/amd64**, it may work on older versions.

It *should work* on all other Go supported OSes.

You can easily install a full node by:

    go get github.com/toqueteos/altcoin/cmd/altcoind

## Organization

Everything lives in its own independent sub-package.

Exceptions to this rule are packages:

- config.
- types.
- tools.
- templates, not an actual Go package; holds the HTML files requires for webgui.

## Comparisons to basiccoin

I've only worried about understanding how this thing worked and making it readable, [basiccoin's Python codebase](https://github.com/zack-bitcoin/basiccoin) was heavily compressed so its very hard to read (even though its Python).

It *may contain* critical bugs due to some missunderstanding while porting it.

**Feel free to criticise, send issues or pull requests at any time.**

I'll try to read all those as soon as possible.

## Creating a derived coin

altcoin allows to easily create derived custom coins easily, just as basiccoin's.

Files you may want to check out:

- [config/config.go](https://github.com/toqueteos/altcoin/blob/master/config/config.go), some generic options like premine, fees, coin name, etc...
- [miner/miner.go](https://github.com/toqueteos/altcoin/blob/master/miner/miner.go) and [miner/pow.go](https://github.com/toqueteos/altcoin/blob/master/miner/pow.go), how the miner works and how Proof-of-Work is implemented.
- [server/server.go](https://github.com/toqueteos/altcoin/blob/master/server/server.go) and [server/request.go](https://github.com/toqueteos/altcoin/blob/master/server/request.go) to customize what `<your-coin-name>d` servers can do.
