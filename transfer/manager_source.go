package transfer

type SourceManager struct {
	packetChan            chan Packet
	signatureChan         chan Checksum

	latestFileInfoPacket  int64
	latestDeltaPacket     int64

	done                  bool
	err                   error

	packeter              *Packeter

	status                *SourceTransferStatus
	stats                 *TransferStats
}


func MakeSourceManager() *SourceManager {

	return &SourceManager{
		packetChan: make(chan Packet, 100),
		signatureChan: make(chan Checksum, SIGNATURE_BUF_SIZE),
		stats: NewTransferStats(),
		packeter: NewPacketer(),
	}
}

// ReceiveStatusUpdate is called by the TCPer.  It's the SourceManager's
// responsibility to call the packeter's "ReceivePacketerStatusUpdate" function
// as well, because the packeter may need to resend some packets, or delete
// some sent packets.
func (manager *SourceManager) ReceiveStatusUpdate (status DestinationTranferStatus) *SourceTransferStatus{

	if status.Failed != nil {
		manager.status.Failed = status.Failed
		manager.ReportError(status.Failed)
	}

	// Tell the packeter about it's counterpart's status. The packeter then
	// return's it's status, which will be sent by the TCPer on it's next iteration.
	manager.status.SourcePacketerStatus = manager.packeter.ReceivePacketerStatusUpdate(
		status.DestinationPacketerStatus)

	// All signature packets have been decoded, call SignatureDone
	if status.LastSignaturePacket != 0 && manager.packeter.LastPacketDecoded >= status.LastSignaturePacket {
		manager.SignatureDone()
	}

	if status.PatchDone {
		manager.PatchDone()
	}

	return manager.status
}


func (manager *SourceManager) QueueFileInfo (fi FileInfo) {
	manager.stats.RecordFileInfo(fi)

	// encoder := ??
	// packets := MakePackets(encoder.MakeBytes(fi), FileInfoPacketType)
	packets := make([]Packet, 1)
	packetNumber, err := manager.packeter.SendPackets(packets)
	if err != nil {
		manager.ReportError(err)
		return
	}

	manager.latestFileInfoPacket = packetNumber

}

func (manager *SourceManager) FileInfoDone () {
	// record the latestFileInfoPacket as the LastFileInfoPacket
	manager.status.LastFileInfoPacket = manager.latestFileInfoPacket
}

func (manager *SourceManager) FileInfoChannel() chan FileInfo {
	return nil
}

func (manager *SourceManager) QueueSignature (sig Checksum) {
	manager.stats.RecordSignature(sig)
	manager.signatureChan <- sig
}

func (manager *SourceManager) SignatureDone() {
	close(manager.signatureChan)
}

func (manager *SourceManager) SignatureChannel() chan Checksum {
	return manager.signatureChan
}

func (manager *SourceManager) QueueDelta (delta Delta) {
	manager.stats.RecordDelta(delta)

	// encoder := ??
	// packets := MakePackets(encoder.MakeBytes(delta), DeltaPacketType)
	packets := make([]Packet, 1)
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

func (manager *SourceManager) ReportError(err error) {
	manager.err = err
}

func (manager *SourceManager) Error() error {
	return manager.err
}

func (manager *SourceManager) Done() bool {
	return manager.done
}

func (manager *SourceManager) Stats() *TransferStats {
	return manager.stats
}
