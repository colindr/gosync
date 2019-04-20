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
func TCPSourceLoop(conn net.Conn, opts *Options, manager *SourceManager){

	// no matter what happens, we close the connection at the end
	defer conn.Close()
	// tell the packeter to close at the end
	defer manager.Packeter().Close()

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	destStatus := &DestinationTransferStatus{}
	sourceStatus := &SourceTransferStatus{}

	sentError := ""

	for (!manager.Done() && sentError=="") {


		//Debug(fmt.Sprintf("Sending sourceStatus %v", sourceStatus))
		// source sends first
		if err :=  encoder.Encode(sourceStatus); err != nil {
			manager.ReportError(err)
			break
		}
		//Debug("Sent sourceStatus.")

		sentError = sourceStatus.Failed

		//Debug("Getting destStatus...")
		// now read the dest's status
		if err := decoder.Decode(destStatus); err != nil {
			manager.ReportError(err)
			break
		}

		//Debug(fmt.Sprintf("Got destStatus %v", destStatus))

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

func TCPDestinationLoop(conn net.Conn, opts *Options, manager *DestinationManager){
	// no matter what happens, we close the connection at the end
	defer conn.Close()
	// tell the packeter to close at the end
	defer manager.Packeter().Close()

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	destStatus := &DestinationTransferStatus{}
	sourceStatus := &SourceTransferStatus{}

	sentError := ""

	for (!manager.status.PatchDone && sentError == ""){

		//Debug("Getting sourceStatus...")
		// destination read's first
		// now read the dest's status
		if err := decoder.Decode(sourceStatus); err != nil {
			manager.ReportError(err)
			break
		}

		//Debug(fmt.Sprintf("Got sourceStatus %v", sourceStatus))

		destStatus = manager.ReceiveStatusUpdate(sourceStatus)

		//Debug(fmt.Sprintf("Sending destStatus %v", destStatus))

		// now send the destination status
		if err :=  encoder.Encode(destStatus); err != nil {
			manager.ReportError(err)
			break
		}
		//Debug("Sent destStatus.")

		sentError = destStatus.Failed

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