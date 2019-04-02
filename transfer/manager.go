package transfer


type Manager interface {
	// QueueFileInfo will queue a FileInfo that will be
	// sent to the signature processor
	QueueFileInfo(fi FileInfo)
	// FileInfoDone should be called when there are no more
	// FileInfos to be generated
	FileInfoDone()
	// FileInfoChannel returns a channel that should be used
	// by the signature processor to read FileInfos.  Will be
	// closed when all FileInfos have been put in the channel.
	FileInfoChannel() chan FileInfo
	// QueueSignature will queue a Checksum that will be
	// sent to the delta processor
	QueueSignature(checksum Checksum)
	// SignatureDone should be called when there are no more
	// Checksums to be generated
	SignatureDone()
	// SignatureChannel returns a channel that should be used
	// by the delta processor to read Checksums.  Will be
	// closed when all Checksums have been put in the channel.
	SignatureChannel() chan Checksum
	// QueueDelta will queue a Delta that will be
	// sent to the patch processor
	QueueDelta(delta Delta)
	// DeltaDone should be called when there are no more
	// Deltas to be generated
	DeltaDone()
	// DeltaChannel returns a channel that should be used
	// by the patch processor to read Delta.  Will be
	// closed when all Deltas have been put in the channel.
	DeltaChannel() chan Delta
	// PatchDone should be called when all deltas have been
	// processed by the patch processor and the transfer is
	// complete.
	PatchDone()
	// ReportError should be called when an error has been
	// reported, it will make sure all channels are closed
	// and it will also make sure that InError() will return
	// True so goroutines will stop doing anything.
	ReportError(err error)
	// Error returns whatever non-nil error that was passed by anyone
	// to ReportError
	Error() error
	// Done returns true PatchDone was called
	Done() bool
	// Stats returns the stats recorded by the manager
	Stats() *TransferStats

}