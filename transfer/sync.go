package transfer

import (
	"net"
	"time"
)


func SyncOutgoing(conn net.Conn, opts *Options, resp *RequestResponse) (*TransferStats, error) {
	// Verify request
	if err := opts.Verify(); err != nil {
		return nil, err
	}

	packeter := NewPacketer()
	manager := NewSourceManager()

	// packet decoder
	go DecodePackets(manager, packeter)

	// tcp loop passes transfer status information between source and dest
	go TCPSourceLoop(conn, opts, packeter, manager)

	// start udp sender gorouting
	go UDPSender(opts.DestinationHost, opts.DestinationUDPPort, opts, packeter, manager)

	// start udp receiver goroutine
	go UDPReceiver(opts.SourceHost, opts.SourceUDPPort, opts, packeter, manager)

	// Outgoing transfer side only does Walk and deltas
	go Walk(opts, manager)
	go ProcessDeltas(opts, manager)

	for {
		if manager.Error() != nil {
			return manager.Stats(), manager.Error()
		} else if manager.Done() {
			return manager.Stats(), nil
		}
		time.Sleep(1)
	}


}

func SyncIncoming(conn net.Conn, opts *Options, resp *RequestResponse) (*TransferStats, error) {
	// Verify request
	if err := opts.Verify(); err != nil {
		return nil, err
	}

	packeter := NewPacketer()
	manager := NewDestinationManager()

	// packet decoder
	go DecodePackets(manager, packeter)

	// tcp loop passes transfer status information between source and dest
	go TCPDestinationLoop(conn, opts, packeter, manager)

	// start udp sender gorouting
	go UDPSender(opts.SourceHost, opts.SourceUDPPort, opts, packeter, manager)

	// start udp receiver goroutine
	go UDPReceiver(opts.DestinationHost, opts.DestinationUDPPort, opts, packeter, manager)

	// Incoming transfer side only does signatures and patches
	go ProcessSignatures(opts, manager)
	go ProcessPatches(opts, manager)

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
func SyncLocal(opts *Options) (*TransferStats, error) {

	// Verify request
	if err := opts.Verify(); err != nil {
		return nil, err
	}

	manager := MakeLocalManager()

	// Super simple
	go Walk(opts, manager)
	go ProcessSignatures(opts, manager)
	go ProcessDeltas(opts, manager)
	go ProcessPatches(opts, manager)


	for {
		if manager.Error() != nil {
			return manager.Stats(), manager.Error()
		} else if manager.Done() {
			return manager.Stats(), nil
		}
		time.Sleep(1)
	}

}