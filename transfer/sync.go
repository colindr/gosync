package transfer

import (
	"errors"
	"fmt"
	"net"
)

func Sync(conn *net.Conn, req *Request, resp *RequestResponse) error {
	if req.Direction == Outgoing {
		return SyncOutgoing(conn, req, resp)
	} else if req.Direction == Incoming {
		return SyncIncoming(conn, req, resp)
	} else {
		return errors.New(fmt.Sprintln("unknown transfer direction:", req.Direction))
	}
}
func SyncOutgoing(conn *net.Conn, req *Request, resp *RequestResponse) error {
    // we send a goroutine to get file information
    // from the filesystem, and send that information via filechan
    return nil
}

func SyncIncoming(conn *net.Conn, req *Request, resp *RequestResponse) error {
	return nil
}
