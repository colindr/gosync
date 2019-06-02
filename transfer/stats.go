package transfer

import (
	"os"
)

type NetStats struct {
	TCPLoopIterations        int64
	ResentSourcePackets      int64
	ResentDestinationPackets int64
}
type TransferStats struct {
	Files         int64
	Symlinks      int64
	Directories   int64
	SourceSize    int64
	BytesSent     int64
	BytesSame     int64
	BytesCopyDest int64
	SigCacheHits  int64
	NetStats      *NetStats
}

func NewTransferStats() *TransferStats {
	return &TransferStats{
		Files:         int64(0),
		Symlinks:      int64(0),
		Directories:   int64(0),
		SourceSize:    int64(0),
		BytesSent:     int64(0),
		BytesCopyDest: int64(0),
		SigCacheHits:  int64(0),
		NetStats: &NetStats{
			TCPLoopIterations:        int64(0),
			ResentSourcePackets:      int64(0),
			ResentDestinationPackets: int64(0),
		},
	}
}

func (s *TransferStats) RecordTCPLoopIteration() {
	s.NetStats.TCPLoopIterations++
}

func (s *TransferStats) RecordResentSourcePackets(n int) {
	s.NetStats.ResentSourcePackets += int64(n)
}

func (s *TransferStats) RecordResentDestinationPackets(n int) {
	s.NetStats.ResentDestinationPackets += int64(n)
}

func (s *TransferStats) RecordFileInfo(fi FileInfo) {
	// count files, directories, symlinks
	if fi.Mode.IsDir() {
		s.Directories += 1
	} else if fi.Mode&os.ModeSymlink == os.ModeSymlink {
		s.Symlinks += 1
	} else {
		s.Files += 1
	}

	// add size to source
	s.SourceSize += fi.Size
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
