package transfer

import (
	"encoding/gob"
	"fmt"
	"net"
	"time"
)

// RequestDone
type RequestDone struct {
	Done bool
}

// The order of operations in finishing a transfer is very specific to
// which side is the source and which is the destination.  The destination
// first calls PatchDone on it's manager, which then sends PatchDone as
// part of the DestinationTransferStatus.  The source then does the normal
// ReceiveTransferStatus and then calls it's own PatchDone, which sets
// manager.done to True.  At this point it's the responsibility of the
// TCPSourceLoop to send RequestDone and read another RequestDone before
// terminating.
func TCPSourceLoop(conn net.Conn, opts *Options, manager *SourceManager) {

	// no matter what happens, we close the connection at the end
	defer conn.Close()
	// tell the manager to close down all remaining net communication, and that
	// this loop is complete
	defer manager.TCPDone()

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	destStatus := DestinationTransferStatus{}
	sourceStatus := SourceTransferStatus{}

	sentError := ""

	for !manager.Done() && sentError == "" {

		manager.stats.RecordTCPLoopIteration()

		Debug(fmt.Sprintf("Sending sourceStatus %v", sourceStatus))
		if err := conn.SetWriteDeadline(time.Now().Add(time.Duration(1) * time.Second)); err != nil {
			manager.ReportError(err)
			break
		}
		// source sends first
		if err := encoder.Encode(&sourceStatus); err != nil {
			manager.ReportError(err)
			break
		}
		Debug("Sent sourceStatus.")

		sentError = sourceStatus.Failed

		Debug("Getting destStatus...")

		if err := conn.SetReadDeadline(time.Now().Add(time.Duration(1) * time.Second)); err != nil {
			manager.ReportError(err)
			break
		}

		// now read the dest's status
		if err := decoder.Decode(&destStatus); err != nil {
			manager.ReportError(err)
			break
		}

		Debug(fmt.Sprintf("Got destStatus %v", destStatus))

		sourceStatus = manager.ReceiveStatusUpdate(destStatus)

		time.Sleep(time.Millisecond * 100)
	}

	if manager.Error() != nil {
		return
	}

	Debug("TCPSourceLoop ending, sending RequestDone...")

	if err := conn.SetWriteDeadline(time.Time{}); err != nil {
		manager.ReportError(err)
		return
	}
	// We must be Done(), so send a RequestDone struct
	if err := encoder.Encode(&RequestDone{Done: true}); err != nil {
		manager.ReportError(err)
		return
	}

	Debug("TCPSourceLoop ending, reading RequestDone...")

	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		manager.ReportError(err)
		return
	}

	reqDone := &RequestDone{}
	// Read one RequestDone
	if err := decoder.Decode(reqDone); err != nil {
		manager.ReportError(err)
		return
	}

	Debug("TCPSourceLoop done")

}

func TCPDestinationLoop(conn net.Conn, opts *Options, manager *DestinationManager) {
	// no matter what happens, we close the connection at the end
	defer conn.Close()
	// tell the manager to close down all remaining net communication, and that
	// this loop is complete
	defer manager.TCPDone()

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	destStatus := DestinationTransferStatus{}
	sourceStatus := SourceTransferStatus{}

	sentError := ""
	sentPatchDone := false

	for !sentPatchDone && sentError == "" {

		manager.stats.RecordTCPLoopIteration()

		Debug("Getting sourceStatus...")
		if err := conn.SetReadDeadline(time.Now().Add(time.Duration(1) * time.Second)); err != nil {
			manager.ReportError(err)
			break
		}
		// destination read's first
		// now read the dest's status
		if err := decoder.Decode(&sourceStatus); err != nil {
			manager.ReportError(err)
			break
		}

		Debug(fmt.Sprintf("Got sourceStatus %v", sourceStatus))

		destStatus = manager.ReceiveStatusUpdate(sourceStatus)

		Debug(fmt.Sprintf("Sending destStatus %v", destStatus))
		if err := conn.SetWriteDeadline(time.Now().Add(time.Duration(1) * time.Second)); err != nil {
			manager.ReportError(err)
			break
		}

		// now send the destination status
		if err := encoder.Encode(&destStatus); err != nil {
			manager.ReportError(err)
			break
		}
		sentPatchDone = destStatus.PatchDone

		Debug(fmt.Sprintf("Sent destStatus %v.", destStatus))

		sentError = destStatus.Failed
		time.Sleep(time.Millisecond * 100)
	}

	if manager.Error() != nil {
		return
	}

	Debug(fmt.Sprintf("TCPDestinationLoop ending %v %v, reading RequestDone...", sentPatchDone, sentError))
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		manager.ReportError(err)
		return
	}
	reqDone := &RequestDone{}
	// We sent a PatchDone status to the source, so we read a RequestDone
	// packet to confirm we're done.
	if err := decoder.Decode(reqDone); err != nil {
		manager.ReportError(err)
		return
	}

	Debug("TCPDestinationLoop ending, sending RequestDone...")
	if err := conn.SetWriteDeadline(time.Time{}); err != nil {
		manager.ReportError(err)
		return
	}

	// And we send just one back
	if err := encoder.Encode(&RequestDone{Done: true}); err != nil {
		manager.ReportError(err)
		return
	}

}
