package transfer

// TODO: remove comm.go, it's now handled by the various manager*.go

// CommArray represents the communication channels necessary
// between fileinfo, signature, delta, and patch goroutines.
// It has an in and an out channel for each part, which allows
// for an intermediary to gather stats and do any necessary
// network communications

type CommArray struct {
	FileInfoIn       chan FileInfo
	FileInfoOut      chan FileInfo
	SignatureIn      chan Checksum
	SignatureOut     chan Checksum
	DeltaIn          chan Delta
	DeltaOut         chan Delta
	ErrorChan        chan error
	DoneChan         chan bool
	Stats            *TransferStats
}




func MakeCommArray() *CommArray {

	return &CommArray{
		FileInfoIn: make(chan FileInfo, FILE_INFO_BUF_SIZE),
		FileInfoOut: make(chan FileInfo, FILE_INFO_BUF_SIZE),
		SignatureIn: make(chan Checksum, SIGNATURE_BUF_SIZE),
		SignatureOut: make(chan Checksum, SIGNATURE_BUF_SIZE),
		DeltaIn: make(chan Delta, DELTA_BUF_SIZE),
		DeltaOut: make(chan Delta, DELTA_BUF_SIZE),
		ErrorChan: make(chan error),
		DoneChan: make(chan bool),
		Stats: NewTransferStats(),
	}
}

func (comm CommArray) FinishFileIn () {
	close(comm.FileInfoIn)
}

func (comm CommArray) FinishSignature() {
	close(comm.SignatureIn)
}

func (comm CommArray) FinishDelta() {
	close(comm.DeltaIn)
}

func (comms CommArray) LocalCommunication() {
	fileInfoInChan := comms.FileInfoIn
	sigInChan := comms.SignatureIn
	deltaInChan := comms.DeltaIn

	for {
		select {
			case fi, ok := <- fileInfoInChan:
				if !ok {
					close(comms.FileInfoOut)
					fileInfoInChan = nil
				} else {
					comms.Stats.RecordFileInfo(fi)
					comms.FileInfoOut <- fi
				}
			case sig, ok := <- sigInChan:
				if !ok {
					close(comms.SignatureOut)
					sigInChan = nil
				} else {
					comms.Stats.RecordSignature(sig)
					comms.SignatureOut <- sig
				}
			case delta, ok := <- deltaInChan:
				if !ok {
					close(comms.DeltaOut)
					deltaInChan = nil
				} else {
					comms.Stats.RecordDelta(delta)
					comms.DeltaOut <- delta
				}
		}

		if fileInfoInChan == nil && sigInChan == nil && deltaInChan == nil {
			break
		}
	}
}

