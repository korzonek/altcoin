// gui.go uses this file to display html to browser.

// +build ignore

package gui

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/toqueteos/altcoin/types"
)

func GET(db *types.DB, request_dict, fn func(db *types.DB)) {
	path := request_dict["uri"] //[1:]
	message = fn(db)

	// return {
	//     "code": "200",
	//     "message": message,
	//     "headers": {
	//         "Content-Type": "text/html",
	//         "Content-Length": str(len(message))
	//     }
	// }
}

// func server(db *types.DB, port, fnGet, fnPost) {
//     func handler(request_dict):
//         method = request_dict["method"]
//         if method == "GET" {
//             return GET(db, request_dict, fnGet)
//         }
//         if method == "POST" {
//             return POST(db, request_dict, fnPost)
//         }
//         return {"code": "501"}  // Method not implemented.

//     //Yashttpd.serve_forever("localhost", port, CONQ, CHUNK, handler)

//     http.Handle("/foo", fooHandler)
//     http.HandleFunc("/bar", )

//     log.Fatal(http.ListenAndServe(":8080", nil))
// }

func Get(rw http.ResponseWriter, req *http.Request) {
	//fmt.Fprintf(rw, "Hello, %q", html.EscapeString(r.URL.Path))
}

func Post(rw http.ResponseWriter, req *http.Request) {
	// func POST(DB, request_dict, fn func(db *types.DB, )) {
	if req.URL.Path != "/home" {
		// return {"code": "404"}
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// var b bytes.Buffer
	// io.Copy(req.Body)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

	}

	values, queryErr := url.ParseQuery(body)
	message := gui.HomePost(db, values)

	// Status code response is automagically set to "200 OK",
	// see ResponseWriter.WriteHeader method.
	fmt.Fprintln(rw, message)
}
