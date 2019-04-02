package transfer

import (
	"hash"
	"io"
	"os"
)

type Checksum struct {
	TransferFile   FileInfo
	SumLen         int
	Sum            hash.Hash
	Len            int
	Offset         int64
	EOF            bool
	Done           bool
}

func ProcessSignatures( req *Request, manager Manager) {

	defer manager.SignatureDone()

	for fileinfo := range manager.FileInfoChannel() {
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
			manager.QueueSignature(c)
			continue

		} else if err != nil {
			// error statting destination
			manager.ReportError(err)
			return

		}

		file, err := os.Open(fileinfo.DestinationPath)
		if err != nil {
			manager.ReportError(err)
			return
		}

		var offset int64
		offset = 0
		buf := make([]byte, req.BlockSize)


		var c Checksum
		var n int


		for {
			n, err = file.Read(buf)

			if err != nil {
				break
			}

			h, sigerr := Signature(buf[:n])
			if sigerr != nil {
				manager.ReportError(sigerr)
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

			manager.QueueSignature(c)

		}

		if err == io.EOF {
			// make a final EOF signature
			c = Checksum{
				TransferFile: fileinfo,
				Len: 0,
				Offset: offset,
				EOF: true,
			}
			manager.QueueSignature(c)

		} else if err != nil {
			manager.ReportError(err)
			return
		}

	}
	return
}