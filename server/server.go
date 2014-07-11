// Sometimes peers ask us for information or push new transactions or blocks to us. This file explains how we respond.
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/toqueteos/altcoin/config"
	"github.com/toqueteos/altcoin/tools"
	"github.com/toqueteos/altcoin/types"
)

const MaxMessageSize = 65536 // 64kb, instead of 60000

var (
	ErrSize = errors.New("Wrong sized message")

	logger = log.New(os.Stdout, "[server] ", log.Ldate|log.Ltime|log.Lshortfile)

	funcs = map[string]func(*Request, *types.DB) *Response{
		"BlockCount":   BlockCount,
		"RangeRequest": RangeRequest,
		"Txs":          Txs,
		"PushTx":       PushTx,
		"PushBlock":    PushBlock,
	}

	// apiCalls = funcs.keys()
	apiCalls = []string{
		"BlockCount",
		"RangeRequest",
		"Txs",
		"PushTx",
		"PushBlock",
	}
)

func SendCommand(peer string, req *Request) (*Response, error) {
	if length := tools.JSONLen(req); length < 1 || length > MaxMessageSize {
		return nil, ErrSize
	}

	conn, err := net.Dial("tcp", peer)
	if err != nil {
		return nil, fmt.Errorf("[server.SendCommand] net.Dial error: %v", err)
	}

	// Write request
	enc := json.NewEncoder(conn)
	if err := enc.Encode(req); err != nil {
		return nil, fmt.Errorf("[server.SendCommand] json.Marshal error: %v", err)
	}

	// Read response back
	dec := json.NewDecoder(conn)
	var resp Response
	if err := dec.Decode(&resp); err != nil {
		return nil, fmt.Errorf("[server.SendCommand] json.Unmarshal error: %v", err)
	}

	return &resp, nil
}

func Run(db *types.DB) {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", config.Get().ListenPort))
	if err != nil {
		logger.Fatalln("net.Listen error:", err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			logger.Println("Couldn't accept client. Error:", err)
			continue
		}

		go Main(conn, db)
	}
}

func Main(conn net.Conn, db *types.DB) {
	var req Request
	dec := json.NewDecoder(conn)
	err := dec.Decode(&req)
	if err != nil {
		logger.Println("Couldn't decode request. Error:", err)
		return
	}

	call := req.Type
	if tools.NotIn(call, apiCalls) {
		logger.Printf("Unknown service: %q\n", call)
	}

	resp := SecurityCheck(&req)
	if !resp.Secure || resp.Error != "ok" {
		logger.Println("SecurityCheck:", resp.Error)
		return
	}

	// try:
	//     return funcs[call](check["newdict"], DB)
	// except:
	//     pass
	fn := funcs[call]
	fn(&req, db)
}
