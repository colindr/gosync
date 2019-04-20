package transfer

import (
	"encoding/gob"
	"fmt"
	"net"
)

func Daemon(addr string) {
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
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	fmt.Println(conn)

	decoder := gob.NewDecoder(conn)
	req := &Request{}

	if err := decoder.Decode(req); err != nil {
		fmt.Println("Error decoding transfer request:", err)
		return
	}

	// TODO: validate request before accepting
	resp := &RequestResponse{
		RequestID: req.RequestID,
		Accepted:  true,
		UDPPort: 30001,  // TODO: identify available port
	}

	encoder := gob.NewEncoder(conn)

	if err := encoder.Encode(resp); err != nil {
		fmt.Println("Error encoding transfer request response:", err)
	}

	opts := &Options{
		Path: req.Path,
		Destination: req.Destination,

		FollowLinks: req.FollowLinks,
		BlockSize: req.BlockSize,
	}

	if req.Direction == Incoming {
		opts.SourceHost = req.Host
		opts.SourceUDPPort = resp.UDPPort

		opts.DestinationHost = req.RequesterHost
		opts.DestinationUDPPort = req.RequesterUDPPort

		SyncOutgoing(conn, opts)
	} else {
		opts.SourceHost = req.RequesterHost
		opts.SourceUDPPort = req.RequesterUDPPort

		opts.DestinationHost = req.Host
		opts.DestinationUDPPort = resp.UDPPort

		SyncIncoming(conn, opts)
	}
}
