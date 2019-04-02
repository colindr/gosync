package transfer

type LocalManager struct {
	fileInfoChan       chan FileInfo
	signatureChan      chan Checksum
	deltaChan          chan Delta

	done               bool
	err                error

	stats            *TransferStats
}

// TODO: make these args?
var FILE_INFO_BUF_SIZE = 10
var SIGNATURE_BUF_SIZE = 10
var DELTA_BUF_SIZE = 10

func MakeLocalManager() *LocalManager {

	return &LocalManager{
		fileInfoChan: make(chan FileInfo, FILE_INFO_BUF_SIZE),
		signatureChan: make(chan Checksum, SIGNATURE_BUF_SIZE),
		deltaChan: make(chan Delta, DELTA_BUF_SIZE),
		stats: NewTransferStats(),
	}
}

func (manager *LocalManager) QueueFileInfo (fi FileInfo) {
	manager.stats.RecordFileInfo(fi)
	manager.fileInfoChan <- fi
}

func (manager *LocalManager) FileInfoDone () {
	close(manager.fileInfoChan)
}

func (manager *LocalManager) FileInfoChannel() chan FileInfo {
	return manager.fileInfoChan
}

func (manager *LocalManager) QueueSignature (sig Checksum) {
	manager.stats.RecordSignature(sig)
	manager.signatureChan <- sig
}

func (manager *LocalManager) SignatureDone() {
	close(manager.signatureChan)
}

func (manager *LocalManager) SignatureChannel() chan Checksum {
	return manager.signatureChan
}

func (manager *LocalManager) QueueDelta (delta Delta) {
	manager.stats.RecordDelta(delta)
	manager.deltaChan <- delta
}

func (manager *LocalManager) DeltaDone() {
	close(manager.deltaChan)
}

func (manager *LocalManager) DeltaChannel() chan Delta {
	return manager.deltaChan
}

func (manager *LocalManager) PatchDone() {
	manager.done = true
}

func (manager *LocalManager) ReportError(err error) {
	manager.err = err
}

func (manager *LocalManager) Error() error {
	return manager.err
}

func (manager *LocalManager) Done() bool {
	return manager.done
}

func (manager *LocalManager) Stats() *TransferStats {
	return manager.stats
}
