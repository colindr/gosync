package transfer

import (
	"encoding/gob"
	"net"
)


// RequestDone
type RequestDone struct {
	Done         bool
}


// The order of operations in finishing a transfer is very specific to
// which side is the source and which is the destination.  The destination
// first calls PatchDone on it's manager, which then sends PatchDone as
// part of the DestinationTransferStatus.  The source then does the normal
// ReceiveTransferStatus and then calls it's own PatchDone, which sets
// manager.done to True.  At this point it's the responsibility of the
// TCPSourceLoop to send RequestDone and read another RequestDone before
// terminating.
func TCPSourceLoop(conn net.Conn, opts *Options, packeter *Packeter, manager *SourceManager){

	// no matter what happens, we close the connection at the end
	defer conn.Close()
	// tell the packeter to close at the end
	defer packeter.Close()

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	var destStatus *DestinationTransferStatus
	var sourceStatus *SourceTransferStatus

	for (!manager.Done() && manager.Error()==nil) {

		// source sends first
		if err :=  encoder.Encode(sourceStatus); err != nil {
			manager.ReportError(err)
			break
		}
		// now read the dest's status
		if err := decoder.Decode(destStatus); err != nil {
			manager.ReportError(err)
			break
		}

		sourceStatus = manager.ReceiveStatusUpdate(destStatus)
	}

	if manager.Error() != nil {
		return
	}

	// We must be Done(), so send a RequestDone struct
	if err := encoder.Encode(&RequestDone{Done:true}); err !=nil{
		manager.ReportError(err)
		return
	}

	var reqDone *RequestDone
	// Read one RequestDone
	if err := decoder.Decode(reqDone); err !=nil{
		manager.ReportError(err)
		return
	}

}

func TCPDestinationLoop(conn net.Conn, opts *Options, packeter *Packeter, manager *DestinationManager){
	// no matter what happens, we close the connection at the end
	defer conn.Close()
	// tell the packeter to close at the end
	defer packeter.Close()

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	var destStatus *DestinationTransferStatus
	var sourceStatus *SourceTransferStatus

	for (!manager.status.PatchDone && manager.Error()==nil){

		// destination read's first
		// now read the dest's status
		if err := decoder.Decode(sourceStatus); err != nil {
			manager.ReportError(err)
			break
		}

		destStatus = manager.ReceiveStatusUpdate(sourceStatus)

		// now send the destination status
		if err :=  encoder.Encode(destStatus); err != nil {
			manager.ReportError(err)
			break
		}

	}

	if manager.Error() != nil {
		return
	}

	var reqDone *RequestDone
	// We sent a PatchDone status to the source, so we read a RequestDone
	// packet to confirm we're done.
	if err := decoder.Decode(reqDone); err !=nil{
		manager.ReportError(err)
		return
	}

	// And we send just one back
	if err := encoder.Encode(&RequestDone{Done:true}); err !=nil{
		manager.ReportError(err)
		return
	}

}