package transfer

type OutgoingManager struct {
	fileInfoChan       chan FileInfo
	signatureChan      chan Checksum
	deltaChan          chan Delta

	done               bool
	err                error

	stats            *TransferStats
}


func MakeOutgoingManager() *OutgoingManager {

	return &OutgoingManager{
		fileInfoChan: make(chan FileInfo, FILE_INFO_BUF_SIZE),
		signatureChan: make(chan Checksum, SIGNATURE_BUF_SIZE),
		deltaChan: make(chan Delta, DELTA_BUF_SIZE),
		stats: NewTransferStats(),
	}
}

func (manager OutgoingManager) QueueFileInfo (fi FileInfo) {
	manager.stats.RecordFileInfo(fi)
	manager.fileInfoChan <- fi
}

func (manager OutgoingManager) FileInfoDone () {
	close(manager.fileInfoChan)
}

func (manager OutgoingManager) FileInfoChannel() chan FileInfo {
	return manager.fileInfoChan
}

func (manager OutgoingManager) QueueSignature (sig Checksum) {
	manager.stats.RecordSignature(sig)
	manager.signatureChan <- sig
}

func (manager OutgoingManager) SignatureDone() {
	close(manager.signatureChan)
}

func (manager OutgoingManager) SignatureChannel() chan Checksum {
	return manager.signatureChan
}

func (manager OutgoingManager) QueueDelta (delta Delta) {
	manager.stats.RecordDelta(delta)
	manager.deltaChan <- delta
}

func (manager OutgoingManager) DeltaDone() {
	close(manager.deltaChan)
}

func (manager OutgoingManager) DeltaChannel() chan Delta {
	return manager.deltaChan
}

func (manager OutgoingManager) PatchDone() {
	manager.done = true
}

func (manager OutgoingManager) ReportError(err error) {
	manager.err = err
}

func (manager OutgoingManager) Error() error {
	return manager.err
}

func (manager OutgoingManager) Done() bool {
	return manager.done
}

func (manager OutgoingManager) Stats() *TransferStats {
	return manager.stats
}
