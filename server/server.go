// Sometimes peers ask us for information or push new transactions or blocks to us. This file explains how we respond.
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/toqueteos/altcoin/config"
	"github.com/toqueteos/altcoin/tools"
	"github.com/toqueteos/altcoin/types"
)

var (
	ErrSize = errors.New("Wrong sized message")

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

const MAX_MESSAGE_SIZE = 65536 // 64kb, instead of 60000

func SendCommand(peer string, req *Request) (*Response, error) {
	if length := tools.JsonLen(req); length < 1 || length > MAX_MESSAGE_SIZE {
		return nil, ErrSize
	}

	conn, err := net.Dial("tcp", peer)
	if err != nil {
		// log.Println("net.Dial error:", err)
		// log.Println("Disconnect.")
		return nil, fmt.Errorf("net.Dial error:", err)
	}

	// Write request
	enc := json.NewEncoder(conn)
	if err := enc.Encode(req); err != nil {
		// log.Println("json.Marshal error:", err)
		// log.Println("Disconnect.")
		return nil, fmt.Errorf("json.Marshal error:", err)
	}

	// Read response back
	dec := json.NewDecoder(conn)
	var resp Response
	if err := dec.Decode(&resp); err != nil {
		// log.Println("json.Unmarshal error:", err)
		// log.Println("Disconnect.")
		return nil, fmt.Errorf("json.Unmarshal error:", err)
	}

	return &resp, nil
}

func Run(db *types.DB) {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", config.Get().ListenPort))
	if err != nil {
		// handle error
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Couldn't accept client. Error:", err)
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
		log.Println("Couldn't decode request. Error:", err)
		return
	}

	call := req.Type
	if tools.NotIn(call, apiCalls) {
		log.Printf("[API Error] Unknown service: %q\n", call)
	}

	resp := SecurityCheck(&req)
	if !resp.Secure || resp.Error != "ok" {
		log.Printf("SecurityCheck:", resp.Error)
		return
	}

	// try:
	//     return funcs[call](check["newdict"], DB)
	// except:
	//     pass
	fn := funcs[call]
	fn(&req, db)
}
