package transfer

import (
	"os"
	"path/filepath"
	"strings"
)


func Walk(opts *Options, manager Manager) {
	// close the channel when we're done
	defer manager.FileInfoDone()

	// our walk func just sends os.FileInfo objects to our channel
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// get source path
		sourceParts := strings.Split(path, opts.Path)
		destPath := filepath.Join(opts.Destination, sourceParts[1])

		t := FileInfo{
			FileInfo: info,
			SourcePath: path,
			DestinationPath: destPath,
		}

		manager.QueueFileInfo(t)
		return nil
	}

	// if it was a walk to remember, and errored, return the error
	if err:= filepath.Walk(opts.Path, walkFunc); err != nil {
		manager.ReportError(err)
		return
	}

	return

}