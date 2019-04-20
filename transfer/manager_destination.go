package transfer

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"runtime/debug"
)

type DestinationManager struct {
	packetChan            chan Packet
	fileInfoChan          chan FileInfo
	deltaChan             chan Delta

	fileInfoClosed        bool
	deltaClosed           bool

	latestSignaturePacket uint64

	done                  bool
	err                   error

	packeter              *Packeter

	status                *DestinationTransferStatus
	stats                 *TransferStats
}


func NewDestinationManager() *DestinationManager {

	return &DestinationManager{
		packetChan: make(chan Packet, 100),
		fileInfoChan: make(chan FileInfo, FILE_INFO_BUF_SIZE),
		deltaChan: make(chan Delta, DELTA_BUF_SIZE),
		status: &DestinationTransferStatus{},
		stats: NewTransferStats(),
		packeter: NewPacketer(),
	}
}

// ReceiveStatusUpdate is called by the TCPer.  It's the DestinationManager's
// responsibility to call the packeter's "ReceivePacketerStatusUpdate" function
// as well, because the packeter may need to resend some packets, or delete
// some sent packets.
func (manager *DestinationManager) ReceiveStatusUpdate (status *SourceTransferStatus) *DestinationTransferStatus{

	if status.Failed != "" {
		manager.status.Failed = status.Failed
		manager.err = errors.New(status.Failed)
	}

	// Tell the packeter about it's counterpart's status. The packeter then
	// return's it's status, which will be sent by the TCPer on it's next iteration.
	manager.status.DestinationPacketerStatus = manager.packeter.ReceivePacketerStatusUpdate(
		status.SourcePacketerStatus)

	// All FileInfo packets have been decoded, call FileInfoDone
	if status.LastFileInfoPacket != 0 &&
		manager.packeter.LastPacketDecoded >= status.LastFileInfoPacket &&
		!manager.fileInfoClosed {
		manager.FileInfoDone()
	}

	// All delta packets have been decoded, call DeltaDone
	if status.LastDeltaPacket != 0 &&
		manager.packeter.LastPacketDecoded >= status.LastDeltaPacket &&
		!manager.deltaClosed{
		manager.DeltaDone()
	}

	return manager.status
}


func (manager *DestinationManager) QueueFileInfo (fi FileInfo) {
	manager.stats.RecordFileInfo(fi)
	manager.fileInfoChan <- fi
}

func (manager *DestinationManager) FileInfoDone () {
	manager.fileInfoClosed = true
	close(manager.fileInfoChan)
}

func (manager *DestinationManager) FileInfoChannel() chan FileInfo {
	return manager.fileInfoChan
}

func (manager *DestinationManager) QueueSignature (sig Checksum) {
	manager.stats.RecordSignature(sig)

	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	if err := encoder.Encode(sig); err != nil {
		manager.ReportError(err)
		return
	}

	packets := MakePackets(&buff, SignaturePacket)
	packetNumber, err := manager.packeter.SendPackets(packets)
	if err != nil {
		manager.ReportError(err)
		return
	}

	manager.latestSignaturePacket = packetNumber
}

func (manager *DestinationManager) SignatureDone() {
	// record the latestFileInfoPacket as the LastFileInfoPacket
	manager.status.LastSignaturePacket = manager.latestSignaturePacket
}

func (manager *DestinationManager) SignatureChannel() chan Checksum {
	return nil
}

func (manager *DestinationManager) QueueDelta (delta Delta) {
	manager.stats.RecordDelta(delta)
	manager.deltaChan <- delta
}

func (manager *DestinationManager) DeltaDone() {
	manager.deltaClosed = true
	close(manager.deltaChan)
}

func (manager *DestinationManager) DeltaChannel() chan Delta {
	return manager.deltaChan
}

func (manager *DestinationManager) PatchDone() {
	manager.status.PatchDone = true
}

func (manager *DestinationManager) Packeter() *Packeter {
	return manager.packeter
}

func (manager *DestinationManager) ReportError(err error) {
	stack := debug.Stack()
	manager.err = fmt.Errorf("%s: %s", stack, err)
	manager.status.Failed = fmt.Sprintf("%s: %s", stack, err)
}

func (manager *DestinationManager) Error() error {
	return manager.err
}

func (manager *DestinationManager) Done() bool {
	return manager.done
}

func (manager *DestinationManager) Stats() *TransferStats {
	return manager.stats
}
