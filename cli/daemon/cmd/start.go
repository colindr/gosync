package cmd

import (
	"encoding/gob"
	"fmt"
	"net"

	"github.com/colindr/gotests/gosync/transfer"
	"github.com/spf13/viper"
)

func StartDaemon() {
	addr := fmt.Sprintf("%v:%v", viper.Get("host"), viper.Get("port"))

	fmt.Println("Listening at", addr)
	// TODO: add tls support
	// config := &tls.Config{
	// 	InsecureSkipVerify: true,
	// }
	// ln, err := tls.Listen("tcp", addr, config)
	ln, err := net.Listen("tcp", addr)

	if err != nil {
		fmt.Println(err)
		return
	}
	listen(ln)
}

func listen(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error while listening:", err)
			continue
		}
		go handleConn(&conn)
	}
}

func handleConn(conn *net.Conn) {
	fmt.Println(conn)
	// TODO: read a TransferRequest then write a TransferRequestResponse
	decoder := gob.NewDecoder(*conn)

	req := &transfer.Request{}

	if err := decoder.Decode(req); err != nil {
		fmt.Println("Error decoding transfer request:", err)
		return
	}
	resp := &transfer.RequestResponse{
		RequestID: req.RequestID,
		Accepted:  true,
	}

	if req.Type == transfer.Incoming {
		// Pick a UDP port to listen on
		resp.UDPPort = 30000
	}

	encoder := gob.NewEncoder(*conn)

	if err := encoder.Encode(resp); err != nil {
		fmt.Println("Error encoding transfer request response:", err)
	}

	transfer.Sync(conn, req, resp)

}
