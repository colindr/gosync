package signature

import (
	"github.com/colindr/gotests/gosync/request"
	"hash"
	"io"
	"os"
)

type Checksum struct {
	TransferFile   request.FileInfo
	SumLen         int
	Sum            hash.Hash
	Len            int
	Offset         int64
	EOF            bool
}

func Process( fileInfoChan <-chan request.FileInfo, signatureChan chan<- Checksum, errChan chan<- error ) {

	defer close (signatureChan)

	for fileinfo := range fileInfoChan {
		var err error

		if fileinfo.FileInfo.IsDir() {
			// It's a directory, we just create the directory and continue
			os.Mkdir(fileinfo.DestinationPath, fileinfo.FileInfo.Mode())

			//TODO: chown and chmod if possible

			continue
		}

		if _, err := os.Stat(fileinfo.DestinationPath); os.IsNotExist(err) {
			// destination does not exist, push an EOF checksum and continue
			c := Checksum{
				TransferFile: fileinfo,
				SumLen: 32,
				Sum: nil,
				Offset: 0,
				Len: 0,
				EOF: true,
			}
			signatureChan <- c
			continue

		} else if err != nil {
			// error statting destination
			errChan <- err
			close(errChan)
			return

		}

		file, err := os.Open(fileinfo.DestinationPath)
		if err != nil {
			errChan <- err
			close(errChan)
			return
		}

		var offset int64
		offset = 0
		buf := make([]byte, 4096)


		var c Checksum
		var n int


		for {
			n, err = file.Read(buf)

			if err != nil {
				break
			}

			h, sigerr := Signature(buf[:n])
			if sigerr != nil {
				errChan <- sigerr
				close(errChan)
				return
			}

			c = Checksum{
				TransferFile: fileinfo,
				SumLen: 32,
				Sum: h,
				Len: n,
				Offset: offset,
			}

			offset += int64(n)

			signatureChan <- c

		}

		if err == io.EOF {
			// make a final EOF signature
			c = Checksum{
				TransferFile: fileinfo,
				Len: 0,
				Offset: offset,
				EOF: true,
			}
			signatureChan <- c

		} else if err != nil {
			errChan <- err
			close(errChan)
			return
		}

	}
	return
}