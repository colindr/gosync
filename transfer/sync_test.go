package transfer

import (
	"bytes"
	"fmt"
	"github.com/colindr/gotests/gosync/request"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

func assertFile(t *testing.T, path string, perm os.FileMode, content string, multiply int) {

}

func TestSyncLocalSimple(t *testing.T) {
	var err error

	N := 50000
	source, err := ioutil.TempDir("/tmp", "gosync.test.")
	if err != nil { panic(err) }
	defer os.RemoveAll(source)

	destination, err := ioutil.TempDir("/tmp", "gosync.test.")
	if err != nil { panic(err) }
	defer os.RemoveAll(destination)

	err = ioutil.WriteFile(path.Join(source, "a"), []byte(strings.Repeat("a", N)), 0755)
	if err != nil { panic(err) }
	err = ioutil.WriteFile(path.Join(source, "b"), []byte(strings.Repeat("b", N)), 0755)
	if err != nil { panic(err) }

	req := &request.Request{
		RequestID: uuid.New(),

		Type: request.Local,

		Path: source,
		Destination: destination,

		FollowLinks: false,
		// TODO: Block Size as part of request?
	}

	err = SyncLocal(req)

	if err != nil {
		t.Error(err)
	}

	acontent, err := ioutil.ReadFile(path.Join(destination, "a"))
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(acontent, []byte(strings.Repeat("a", N))) {
		t.Error(fmt.Sprintf("output %s was wrong",path.Join(destination, "a") ))
	}

	bcontent, err := ioutil.ReadFile(path.Join(destination, "b"))
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(bcontent, []byte(strings.Repeat("b", N))) {
		t.Error(fmt.Sprintf("output %s was wrong",path.Join(destination, "b") ))
	}

}