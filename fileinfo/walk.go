package fileinfo

import (
	"github.com/colindr/gotests/gosync/request"
	"os"
	"path/filepath"
	"strings"
)


func Walk(root string, destRoot string, fileinfochan chan<- request.FileInfo, errChan chan<- error) {
	// close the channel when we're done
	defer close(fileinfochan)

	// our walk func just sends os.FileInfo objects to our channel
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// get source path
		sourceParts := strings.Split(path, root)
		destPath := filepath.Join(destRoot, sourceParts[1])

		t := request.FileInfo{
			FileInfo: info,
			SourcePath: path,
			DestinationPath: destPath,
		}

		fileinfochan <- t
		return nil
	}

	// if it was a walk to remember, and errored, return the error
	if err:= filepath.Walk(root, walkFunc); err != nil {
		errChan <- err
		close(errChan)
		return
	}
	return

}