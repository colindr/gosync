package transfer

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"runtime/debug"
)

type SourceManager struct {
	packetChan    chan Packet
	signatureChan chan Checksum

	signatureClosed bool

	latestFileInfoPacket uint64
	latestDeltaPacket    uint64

	done    bool
	tcpdone bool
	err     error

	packeter *Packeter

	status *SourceTransferStatus
	stats  *TransferStats
}

func NewSourceManager() *SourceManager {

	return &SourceManager{
		packetChan:    make(chan Packet, 100),
		signatureChan: make(chan Checksum, SIGNATURE_BUF_SIZE),
		stats:         NewTransferStats(),
		status:        &SourceTransferStatus{},
		packeter:      NewPacketer(),
	}
}

// ReceiveStatusUpdate is called by the TCPer.  It's the SourceManager's
// responsibility to call the packeter's "ReceivePacketerStatusUpdate" function
// as well, because the packeter may need to resend some packets, or delete
// some sent packets.
func (manager *SourceManager) ReceiveStatusUpdate(status DestinationTransferStatus) SourceTransferStatus {

	if status.Failed != "" {
		manager.status.Failed = status.Failed
		manager.err = errors.New(status.Failed)
	}

	// Record the number of packets the destination is requesting to be resent
	manager.stats.RecordResentDestinationPackets(
		len(status.DestinationPacketerStatus.ResendPackets))

	// Tell the packeter about it's counterpart's status. The packeter then
	// return's it's status, which will be sent by the TCPer on it's next iteration.
	manager.status.SourcePacketerStatus = manager.packeter.ReceivePacketerStatusUpdate(
		status.DestinationPacketerStatus)

	// Record the number of packets the source is requesting to be resent
	manager.stats.RecordResentSourcePackets(
		len(manager.status.SourcePacketerStatus.ResendPackets))

	// All signature packets have been decoded, call SignatureDone
	if status.LastSignaturePacket != 0 &&
		manager.packeter.LastPacketDecoded >= status.LastSignaturePacket &&
		!manager.signatureClosed {
		manager.SignatureDone()
	}

	if status.PatchDone {
		manager.PatchDone()
	}

	return *manager.status
}

func (manager *SourceManager) QueueFileInfo(fi FileInfo) {
	manager.stats.RecordFileInfo(fi)

	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	if err := encoder.Encode(fi); err != nil {
		manager.ReportError(err)
		return
	}

	packets := MakePackets(&buff, FileInfoPacket)
	packetNumber, err := manager.packeter.SendPackets(packets)
	if err != nil {
		manager.ReportError(err)
		return
	}

	manager.latestFileInfoPacket = packetNumber

}

func (manager *SourceManager) FileInfoDone() {
	// record the latestFileInfoPacket as the LastFileInfoPacket
	manager.status.LastFileInfoPacket = manager.latestFileInfoPacket
}

func (manager *SourceManager) FileInfoChannel() chan FileInfo {
	return nil
}

func (manager *SourceManager) QueueSignature(sig Checksum) {
	manager.stats.RecordSignature(sig)
	manager.signatureChan <- sig
}

func (manager *SourceManager) SignatureDone() {
	manager.signatureClosed = true
	close(manager.signatureChan)
}

func (manager *SourceManager) SignatureChannel() chan Checksum {
	return manager.signatureChan
}

func (manager *SourceManager) QueueDelta(delta Delta) {
	manager.stats.RecordDelta(delta)

	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	if err := encoder.Encode(delta); err != nil {
		manager.ReportError(err)
		return
	}

	packets := MakePackets(&buff, DeltaPacket)
	packetNumber, err := manager.packeter.SendPackets(packets)
	if err != nil {
		manager.ReportError(err)
		return
	}

	manager.latestDeltaPacket = packetNumber
}

func (manager *SourceManager) DeltaDone() {
	// record the latestDeltaPacket as the LastDeltaPacket
	manager.status.LastDeltaPacket = manager.latestDeltaPacket
}

func (manager *SourceManager) DeltaChannel() chan Delta {
	return nil
}

func (manager *SourceManager) PatchDone() {
	manager.done = true
}

func (manager *SourceManager) Packeter() *Packeter {
	return manager.packeter
}

func (manager *SourceManager) TCPDone() {
	manager.packeter.Close()
	manager.tcpdone = true
}

func (manager *SourceManager) ReportError(err error) {
	Debug(fmt.Sprintf("Error reported: %v", err))
	stack := debug.Stack()
	manager.err = fmt.Errorf("%s: %s", stack, err)
	manager.status.Failed = fmt.Sprintf("%s: %s", stack, err)

}

func (manager *SourceManager) Error() error {
	return manager.err
}

func (manager *SourceManager) Done() bool {
	return manager.done
}

func (manager *SourceManager) NetDone() bool {
	return manager.packeter.Done() && manager.tcpdone
}

func (manager *SourceManager) Stats() *TransferStats {
	return manager.stats
}
