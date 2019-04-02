package transfer

import (
	"fmt"
	"os"
)

func ProcessPatches( req *Request, manager Manager) {
	defer manager.PatchDone()

	fdmap := make(map[string]*os.File)

	for delta := range manager.DeltaChannel() {
		if delta.NoOp {
			Debug(fmt.Sprintf("not touching %s at offset %d\n", delta.Path, delta.Offset))
			continue
		}

		var f *os.File
		f, ok := fdmap[delta.Path]
		if (!ok) {
			openf, err := os.OpenFile(delta.Path, os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				manager.ReportError(err)
				return
			}
			fdmap[delta.Path] = openf
			f = openf
		}

		if delta.EOF {
			if err := f.Truncate(delta.Offset); err != nil {
				manager.ReportError(err)
				return
			}

			if err := f.Sync(); err != nil {
				manager.ReportError(err)
				return
			}

			if err := f.Close(); err != nil {
				manager.ReportError(err)
				return
			}

			continue
		}

		if _, err := f.Seek(delta.Offset, 0); err != nil {
			manager.ReportError(err)
			return
		}

		if _, err := f.Write(delta.Content); err != nil {
			manager.ReportError(err)
			return
		}

		Debug(fmt.Sprintf("( %d %s ) %s\n", delta.Offset, delta.Path, delta.Content))

	}

}
