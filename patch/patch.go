package patch

import (
	"fmt"
	"github.com/colindr/gotests/gosync/delta"
	"github.com/colindr/gotests/gosync/log"
	"github.com/colindr/gotests/gosync/request"
	"os"
)

func Apply( req request.Request, deltaChan <-chan delta.Delta, doneChan chan<- bool, errChan chan<- error) {

	finish := func () { doneChan <- true }

	defer finish()

	fdmap := make(map[string]*os.File)

	for delta := range deltaChan {
		if delta.NoOp {
			log.Debug(fmt.Sprintf("not touching %s at offset %d\n", delta.Path, delta.Offset))
			continue
		}

		var f *os.File
		f, ok := fdmap[delta.Path]
		if (!ok) {
			openf, err := os.OpenFile(delta.Path, os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				errChan <- err
				close(errChan)
				return
			}
			fdmap[delta.Path] = openf
			f = openf
		}

		if delta.EOF {
			if err := f.Truncate(delta.Offset); err != nil {
				errChan <- err
				close(errChan)
				return
			}

			if err := f.Sync(); err != nil {
				errChan <- err
				close(errChan)
				return
			}

			if err := f.Close(); err != nil {
				errChan <- err
				close(errChan)
				return
			}

			continue
		}

		if _, err := f.Seek(delta.Offset, 0); err != nil {
			errChan <- err
			close(errChan)
			return
		}

		if _, err := f.Write(delta.Content); err != nil {
			errChan <- err
			close(errChan)
			return
		}

		log.Debug(fmt.Sprintf("( %d %s ) %s\n", delta.Offset, delta.Path, delta.Content))

	}

}
