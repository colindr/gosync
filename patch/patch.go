package patch

import (
	"fmt"
	"github.com/colindr/gotests/gosync/delta"
	"github.com/colindr/gotests/gosync/request"
	"os"
)

func Apply( req request.Request, deltaChan <-chan delta.Delta, doneChan chan<- bool, errChan chan<- error) {

	finish := func () { doneChan <- true }

	defer finish()

	fdmap := make(map[string]*os.File)

	for delta := range deltaChan {
		if delta.NoOp {
			fmt.Printf("not touching %s at offset %d\n", delta.Path, delta.Offset)
			continue
		}

		var f *os.File
		f, ok := fdmap[delta.Path]
		if (!ok) {
			openf, err := os.Open(delta.Path)
			if os.IsNotExist(err) {
				createf, createerr := os.Create(delta.Path)

				if createerr != nil {
					errChan <- createerr
					close(errChan)
					return
				}
				openf = createf

			} else if err != nil {
				errChan <- err
				close(errChan)
				return
			}
			fdmap[delta.Path] = openf
			f = openf
		}

		f.Write(delta.Content)

		fmt.Printf("wrote %s to %s at offset %d\n", delta.Content, delta.Path, delta.Offset)

	}

}
