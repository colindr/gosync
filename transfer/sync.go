package transfer

import (
	"errors"
	"fmt"
	"net"
	"time"
)

func Sync(conn *net.Conn, req *Request, resp *RequestResponse) error {
	if req.Type == Outgoing {
		_, err := SyncOutgoing(conn, req, resp)
		return err
	} else if req.Type == Incoming {
		_, err := SyncIncoming(conn, req, resp)
		return err
	} else {
		return errors.New(fmt.Sprintln("unknown transfer direction:", req.Type))
	}
}

func SyncOutgoing(conn *net.Conn, req *Request, resp *RequestResponse) (*TransferStats, error) {
	// Verify request
	if err := req.Verify(); err != nil {
		return nil, err
	}

	manager := MakeOutgoingManager()

	// Outgoing transfer side only does Walk and deltas
	go Walk(req, manager)
	go ProcessDeltas(req, manager)

	for {
		if manager.Error() != nil {
			return manager.Stats(), manager.Error()
		} else if manager.Done() {
			return manager.Stats(), nil
		}
		time.Sleep(1)
	}


}

func SyncIncoming(conn *net.Conn, req *Request, resp *RequestResponse) (*TransferStats, error) {
	// Verify request
	if err := req.Verify(); err != nil {
		return nil, err
	}

	manager := MakeIncomingManager()

	// Incoming transfer side only does signatures and patches
	go ProcessSignatures(req, manager)
	go ProcessPatches(req, manager)

	for {
		if manager.Error() != nil {
			return manager.Stats(), manager.Error()
		} else if manager.Done() {
			return manager.Stats(), nil
		}
		time.Sleep(1)
	}

}

// SyncLocal does all filesystem operations locally
func SyncLocal(req *Request) (*TransferStats, error) {

	// Verify request
	if err := req.Verify(); err != nil {
		return nil, err
	}

	manager := MakeLocalManager()

	// Super simple
	go Walk(req, manager)
	go ProcessSignatures(req, manager)
	go ProcessDeltas(req, manager)
	go ProcessPatches(req, manager)


	for {
		if manager.Error() != nil {
			return manager.Stats(), manager.Error()
		} else if manager.Done() {
			return manager.Stats(), nil
		}
		time.Sleep(1)
	}

}