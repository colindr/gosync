package transfer

type IncomingManager struct {
	fileInfoChan       chan FileInfo
	signatureChan      chan Checksum
	deltaChan          chan Delta

	done               bool
	err                error

	stats            *TransferStats
}


func MakeIncomingManager() *IncomingManager {

	return &IncomingManager{
		fileInfoChan: make(chan FileInfo, FILE_INFO_BUF_SIZE),
		signatureChan: make(chan Checksum, SIGNATURE_BUF_SIZE),
		deltaChan: make(chan Delta, DELTA_BUF_SIZE),
		stats: NewTransferStats(),
	}
}

func (manager IncomingManager) QueueFileInfo (fi FileInfo) {
	manager.stats.RecordFileInfo(fi)
	manager.fileInfoChan <- fi
}

func (manager IncomingManager) FileInfoDone () {
	close(manager.fileInfoChan)
}

func (manager IncomingManager) FileInfoChannel() chan FileInfo {
	return manager.fileInfoChan
}

func (manager IncomingManager) QueueSignature (sig Checksum) {
	manager.stats.RecordSignature(sig)
	manager.signatureChan <- sig
}

func (manager IncomingManager) SignatureDone() {
	close(manager.signatureChan)
}

func (manager IncomingManager) SignatureChannel() chan Checksum {
	return manager.signatureChan
}

func (manager IncomingManager) QueueDelta (delta Delta) {
	manager.stats.RecordDelta(delta)
	manager.deltaChan <- delta
}

func (manager IncomingManager) DeltaDone() {
	close(manager.deltaChan)
}

func (manager IncomingManager) DeltaChannel() chan Delta {
	return manager.deltaChan
}

func (manager IncomingManager) PatchDone() {
	manager.done = true
}

func (manager IncomingManager) ReportError(err error) {
	manager.err = err
}

func (manager IncomingManager) Error() error {
	return manager.err
}

func (manager IncomingManager) Done() bool {
	return manager.done
}

func (manager IncomingManager) Stats() *TransferStats {
	return manager.stats
}
