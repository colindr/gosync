package transfer

import (
	"bytes"
	"io"
	"os"
)

// Delta can be applied to the basis file to produce the desired
// result file
type Delta struct {
	Basis   os.FileInfo
	Path    string
	Len     int
	Content []byte
	Offset  int64
	EOF     bool
	NoOp    bool
	Done    bool
}


func ProcessDeltas( req *Request, manager Manager)  {

	defer manager.DeltaDone()

	fdmap := make(map[string]*os.File)

	eofmap := make(map[string]int64)

	buf := make([]byte, req.BlockSize)

	for sig := range manager.SignatureChannel() {
		var f *os.File
		f, ok := fdmap[sig.TransferFile.SourcePath]
		if (!ok) {
			openf, err := os.Open(sig.TransferFile.SourcePath)
			if (err != nil) {
				manager.ReportError(err)
				return
			}
			fdmap[sig.TransferFile.SourcePath] = openf
			f = openf
		}

		if fileEOF, ok := eofmap[sig.TransferFile.SourcePath]; ok {
			// we've already hit the end of this file, don't make any
			// more deltas
			if sig.Offset >= fileEOF {
				continue
			}
		}

		if (sig.EOF) {
			// This signature represents the end of the file
			// so we should generate deltas for the rest of the file after
			// signature.Offset and then close the file

			offset := sig.Offset

			var err error
			var n   int
			for {
				// seek in the source file
				if _, err := f.Seek(offset, 0); err != nil{
					manager.ReportError(err)
					return
				}

				n, err = f.Read(buf)
				// read until we error or get an EOF
				if err != nil {
					break
				}

				// make more copy deltas
				manager.QueueDelta(makeCopyDelta(sig, buf, n, offset))

				offset += int64(n)

			}

			if err != io.EOF {
				manager.ReportError(err)
				return
			}

			// make EOF delta
			manager.QueueDelta(makeEOFDelta(sig, offset))
			// done with this EOF sig, continue getting more sigs
			continue
		}

		// seek in the source file
		if _, err := f.Seek(sig.Offset, 0); err != nil{
			manager.ReportError(err)
			return
		}

		// We don't want to READ *more* bytes than the sig.Len
		buf = buf[:sig.Len]

		n, err := f.Read(buf)

		if err==io.EOF {
			eofmap[sig.TransferFile.SourcePath] = sig.Offset
			manager.QueueDelta(makeEOFDelta(sig, sig.Offset))
		} else if err != nil {
			manager.ReportError(err)
			return
		}

		if n != sig.Len {
			// sizes don't match, just return a copy
			manager.QueueDelta(makeCopyDelta(sig, buf, n, sig.Offset))
		} else {
			h, err := Signature(buf[:n])

			if err != nil {
				manager.ReportError(err)
				return
			}

			if !bytes.Equal(h.Sum(nil), sig.Sum.Sum(nil)) {
				// sizes don't match, just return a dumb_delta
				manager.QueueDelta(makeCopyDelta(sig, buf, n, sig.Offset))
			} else {
				manager.QueueDelta(makeNoCopyDelta(sig))
			}
		}

		if n < len(buf) {
			// We've probably reached the end of the file, so
			// make a EOFDelta and record the end
			eofmap[sig.TransferFile.SourcePath] = sig.Offset + int64(n)
			manager.QueueDelta(makeEOFDelta(sig, sig.Offset + int64(n)))
		}

	}

	return
}

func makeCopyDelta(sig Checksum, buf []byte, length int, offset int64) Delta {

	// Need to make newbuf because buf will be overwritten soon
	newbuf := make([]byte, length)
	copy(newbuf, buf[:length])

	b := Delta{
		Basis: sig.TransferFile.FileInfo,
		Path:  sig.TransferFile.DestinationPath,
		Len: length,
		Content: newbuf[:length],
		Offset: offset,
		NoOp: false,
	}

	return b

}


func makeEOFDelta(sig Checksum, offset int64) Delta {
	b := Delta{
		Basis: sig.TransferFile.FileInfo,
		Path:  sig.TransferFile.DestinationPath,
		Len: 0,
		Offset: offset,
		EOF: true,
	}

	return b
}

func makeNoCopyDelta(sig Checksum) Delta {

	b := Delta{
		Basis: sig.TransferFile.FileInfo,
		Path:  sig.TransferFile.DestinationPath,
		Len: sig.Len,
		Offset: sig.Offset,
		NoOp: true,
	}

	return b

}