package delta

import (
	"bytes"
	"github.com/colindr/gotests/gosync/request"
	"github.com/colindr/gotests/gosync/signature"
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
}


func Process( req request.Request, signatureChan <-chan signature.Checksum,
	deltaChan chan<- Delta, errChan chan<- error , search bool )  {

	defer close (deltaChan)

	fdmap := make(map[string]*os.File)

	eofmap := make(map[string]int64)

	buf := make([]byte, 4096)

	for sig := range signatureChan {
		var f *os.File
		f, ok := fdmap[sig.TransferFile.SourcePath]
		if (!ok) {
			openf, err := os.Open(sig.TransferFile.SourcePath)
			if (err != nil) {
				errChan <- err
				close(errChan)
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
		// seek in the source file
		if _, err := f.Seek(sig.Offset, 0); err != nil{
			errChan <- err
			close(errChan)
			return
		}

		if (sig.EOF) {
			// This signature represents the end of the file
			// so we should generate deltas for the rest of the file after
			// signature.Offset and then close the file

			offset := sig.Offset

			var err error
			var n   int
			for {
				n, err = f.Read(buf)
				// read until we error or get an EOF
				if err != nil {
					break
				}

				// make more copy deltas
				deltaChan <- makeCopyDelta(sig, buf, n, offset)

				offset += int64(n)

			}

			if err != io.EOF {
				errChan <- err
				close(errChan)
				return
			}

			// make EOF delta
			deltaChan <- makeEOFDelta(sig, offset)
			// done with this EOF sig, continue getting more sigs
			continue

		}

		n, err := f.Read(buf)

		if err==io.EOF {
			eofmap[sig.TransferFile.SourcePath] = sig.Offset
			deltaChan <- makeEOFDelta(sig, sig.Offset)
		} else if err != nil {
			errChan <- err
			close(errChan)
			return
		}

		if n != sig.Len {
			// sizes don't match, just return a copy
			deltaChan <- makeCopyDelta(sig, buf, n, sig.Offset)
		} else {
			h, err := signature.Signature(buf[:n])

			if err != nil {
				errChan <- err
				close(errChan)
				return
			}

			if !bytes.Equal(h.Sum(nil), sig.Sum.Sum(nil)) {
				// sizes don't match, just return a dumb_delta
				deltaChan <- makeCopyDelta(sig, buf, n, sig.Offset)
			} else {
				deltaChan <- makeNoCopyDelta(sig)
			}
		}

		if n < len(buf) {
			// We've probably reached the end of the file, so
			// make a EOFDelta and record the end
			eofmap[sig.TransferFile.SourcePath] = sig.Offset + int64(n)
			deltaChan <- makeEOFDelta(sig, sig.Offset + int64(n))
		}

	}

	return
}

func makeCopyDelta(sig signature.Checksum, buf []byte, length int, offset int64) Delta {

	b := Delta{
		Basis: sig.TransferFile.FileInfo,
		Path:  sig.TransferFile.DestinationPath,
		Len: length,
		Content: buf[:length],
		Offset: offset,
		NoOp: false,
	}

	return b

}


func makeEOFDelta(sig signature.Checksum, offset int64) Delta {
	b := Delta{
		Basis: sig.TransferFile.FileInfo,
		Path:  sig.TransferFile.DestinationPath,
		Len: 0,
		Offset: offset,
		EOF: true,
	}

	return b
}

func makeNoCopyDelta(sig signature.Checksum) Delta {

	b := Delta{
		Basis: sig.TransferFile.FileInfo,
		Path:  sig.TransferFile.DestinationPath,
		Len: sig.Len,
		Offset: sig.Offset,
		NoOp: true,
	}

	return b

}