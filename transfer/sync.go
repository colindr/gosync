package transfer

import (
	"errors"
	"fmt"
	"github.com/colindr/gotests/gosync/delta"
	"github.com/colindr/gotests/gosync/fileinfo"
	"github.com/colindr/gotests/gosync/patch"
	"github.com/colindr/gotests/gosync/request"
	"github.com/colindr/gotests/gosync/signature"
	"net"
)

func Sync(conn *net.Conn, req *request.Request, resp *request.RequestResponse) error {
	if req.Type == request.Outgoing {
		return SyncOutgoing(conn, req, resp)
	} else if req.Type == request.Incoming {
		return SyncIncoming(conn, req, resp)
	} else {
		return errors.New(fmt.Sprintln("unknown transfer direction:", req.Type))
	}
}

func SyncOutgoing(conn *net.Conn, req *request.Request, resp *request.RequestResponse) error {
    // we send a goroutine to get file information
    // from the filesystem, and send that information via filechan
    return nil
}

func SyncIncoming(conn *net.Conn, req *request.Request, resp *request.RequestResponse) error {
	return nil
}

// SyncLocal does all filesystem operations locally
func SyncLocal(req *request.Request) error {

	fileinfochan := make(chan request.FileInfo)
	checksumchan := make(chan signature.Checksum)
	deltachan := make(chan delta.Delta)
	errchan := make(chan error)
	donechan := make(chan bool)

	// Super simple
	go fileinfo.Walk(req.Path, req.Destination, fileinfochan, errchan)
	go signature.Process(fileinfochan, checksumchan, errchan)
	go delta.Process(*req, checksumchan, deltachan, errchan, false)
	go patch.Apply(*req, deltachan, donechan, errchan)


	select {
	case err := <-errchan:
		return err
	case  <-donechan:
		return nil
	}

}