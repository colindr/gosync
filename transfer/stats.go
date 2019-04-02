package transfer

import (
	"os"
	"syscall"
)

type TransferStats struct{
	Files          int64
	Symlinks       int64
	Directories    int64
	SourceSize     int64
	BytesSent      int64
	BytesSame      int64
	BytesCopyDest  int64
	SigCacheHits   int64
}

func NewTransferStats() *TransferStats {
	return &TransferStats{
		Files: int64(0),
		Symlinks: int64(0),
		Directories: int64(0),
		SourceSize: int64(0),
		BytesSent: int64(0),
		BytesCopyDest: int64(0),
		SigCacheHits: int64(0),
	}
}

func (s *TransferStats) RecordFileInfo(fi FileInfo) {
	// count files, directories, symlinks
	if fi.FileInfo.IsDir() {
		s.Directories += 1
	} else if fi.FileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		s.Symlinks += 1
	} else {
		s.Files += 1
	}

	// add size to source
	if stat, ok := fi.FileInfo.Sys().(*syscall.Stat_t);ok {
		s.SourceSize += stat.Size
	}
}

func (s *TransferStats) RecordSignature(sig Checksum) {

}

func (s *TransferStats) RecordDelta(delta Delta) {
	s.BytesSent += int64(len(delta.Content))

	if delta.Len != len(delta.Content) {
		s.BytesSame += int64(delta.Len)
	}

	//TODO: implement signature cache and copydest
}