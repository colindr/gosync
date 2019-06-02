package transfer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"testing"
)

type SyncTestCaseFilePiece struct {
	Character rune
	Num       int
}

type SyncTestCaseFile struct {
	RelPath string
	Mode    os.FileMode
	Pieces  []SyncTestCaseFilePiece
}

type SyncTestCase struct {
	SourceFiles []SyncTestCaseFile
	DestFiles   []SyncTestCaseFile
	BlockSize   int
	BytesSent   int64
	BytesSame   int64
	Files       int64
	Directories int64
}

var testcasebasic = SyncTestCase{
	SourceFiles: []SyncTestCaseFile{
		{
			RelPath: "a",
			Pieces: []SyncTestCaseFilePiece{
				{
					Character: 'a',
					Num:       10,
				},
			},
		},
		{
			RelPath: "b",
			Pieces: []SyncTestCaseFilePiece{
				{
					Character: 'b',
					Num:       20,
				},
			},
		},
	},
	DestFiles:   []SyncTestCaseFile{},
	BlockSize:   10,
	BytesSent:   30,
	BytesSame:   0,
	Directories: 1,
	Files:       2,
}

var testcasechecksum = SyncTestCase{
	SourceFiles: []SyncTestCaseFile{
		{
			RelPath: "a",
			Pieces: []SyncTestCaseFilePiece{
				{
					Character: 'a',
					Num:       20,
				},
			},
		},
	},
	DestFiles: []SyncTestCaseFile{
		{
			RelPath: "a",
			Pieces: []SyncTestCaseFilePiece{
				{
					Character: 'a',
					Num:       10,
				},
			},
		},
	},
	BlockSize:   10,
	BytesSent:   10,
	BytesSame:   10,
	Directories: 1,
	Files:       1,
}

func TestAbsPathVerify(t *testing.T) {
	opts := &Options{
		Path:        "a",
		Destination: "/b",

		FollowLinks: false,
		BlockSize:   10,
	}

	_, err := SyncLocal(opts)

	if err == nil {
		t.Error("Should have gotten an absolute path error")
	}

	opts.Path = "/a"
	opts.Destination = "b"

	_, err = SyncLocal(opts)

	if err == nil {
		t.Error("Should have gotten an absolute path error")
	}

	opts.Path = "a"
	opts.Destination = "b"

	_, err = SyncLocal(opts)

	if err == nil {
		t.Error("Should have gotten an absolute path error")
	}
}

func TestBasicLocal(t *testing.T) {
	testcase := testcasebasic
	buildAndRunLocalSyncTest(t, testcase)
}

func TestBasicLocalLargeBlocksize(t *testing.T) {
	testcase := testcasebasic
	// run the same test with a block size larger than the file size
	testcase.BlockSize = 4096
	buildAndRunLocalSyncTest(t, testcase)
}

func TestBasicNet(t *testing.T) {
	testcase := testcasebasic
	buildAndRunNetSyncTest(t, testcase)
}

func TestBasicNetLargeBlocksize(t *testing.T) {
	testcase := testcasebasic
	// run the same test with a block size larger than the file size
	testcase.BlockSize = 4096
	buildAndRunNetSyncTest(t, testcase)
}

func TestChecksumLocal(t *testing.T) {
	testcase := testcasechecksum
	buildAndRunLocalSyncTest(t, testcase)

	// run the same test with a block size larger than the file size
	testcase.BlockSize = 100
	buildAndRunLocalSyncTest(t, testcase)
}

func TestChecksumNet(t *testing.T) {
	testcase := testcasechecksum
	buildAndRunNetSyncTest(t, testcase)
}

func TestChecksumNetLargeBlocksize(t *testing.T) {
	testcase := testcasechecksum
	// run the same test with a block size larger than the file size
	testcase.BlockSize = 100
	buildAndRunNetSyncTest(t, testcase)
}

func assertFiles(t *testing.T, stats *TransferStats, files []SyncTestCaseFile, dir string) {

	numFiles := 0
	for _, f := range files {
		numFiles += 1
		filepath := path.Join(dir, f.RelPath)
		actualcontent, err := ioutil.ReadFile(filepath)
		if err != nil {
			t.Error(err)
		}

		expected := ""
		for _, p := range f.Pieces {
			expected += strings.Repeat(string(p.Character), p.Num)
		}

		if !bytes.Equal(actualcontent, []byte(expected)) {
			t.Error(fmt.Sprintf("output %s was wrong", filepath))
		}
	}

	if stats.Files != int64(numFiles) {
		t.Error(
			fmt.Sprintf("reported stats.Files should be %v instead of %v",
				numFiles,
				stats.Files))
	}
}

func makeFiles(files []SyncTestCaseFile, dir string) {
	for _, f := range files {

		s := ""
		for _, p := range f.Pieces {
			s += strings.Repeat(string(p.Character), p.Num)
		}

		// if mode is not specified, set to 0770
		if f.Mode == 0 {
			f.Mode = 0770
		}
		err := ioutil.WriteFile(path.Join(dir, f.RelPath), []byte(s), f.Mode)
		if err != nil {
			panic(err)
		}
	}
}

// buildAndRunLocalSyncTest is a helper function that
func buildAndRunLocalSyncTest(t *testing.T, testcase SyncTestCase) *TransferStats {

	source, err := ioutil.TempDir("/tmp", "gosync.source.")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(source)

	destination, err := ioutil.TempDir("/tmp", "gosync.dest.")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(destination)

	makeFiles(testcase.SourceFiles, source)
	makeFiles(testcase.DestFiles, destination)

	opts := &Options{

		Path:        source,
		Destination: destination,

		FollowLinks: false,
		BlockSize:   testcase.BlockSize,
	}

	stats, err := SyncLocal(opts)

	if err != nil {
		t.Error(err)
	}

	assertFiles(t, stats, testcase.SourceFiles, destination)

	if stats.BytesSent != testcase.BytesSent {
		t.Error(fmt.Sprintf("BytesSent should have been %v not %v",
			testcase.BytesSent, stats.BytesSent))
	}
	if stats.BytesSame != testcase.BytesSame {
		t.Error(fmt.Sprintf("BytesSame should have been %v not %v",
			testcase.BytesSame, stats.BytesSame))
	}
	if stats.Files != testcase.Files {
		t.Error(fmt.Sprintf("Files should have been %v not %v",
			testcase.Files, stats.Files))
	}
	if stats.Directories != testcase.Directories {
		t.Error(fmt.Sprintf("Directories should have been %v not %v",
			testcase.Directories, stats.Directories))
	}
	return stats
}

func buildAndRunNetSyncTest(t *testing.T, testcase SyncTestCase) *TransferStats {
	// The trick here is to start both sides of a TCP connection, and pass each
	// side to their various SyncOutgoing and SyncIncoming functions.

	// We run both sides on the same host so we're treating the filepaths
	// the same way we do when we're running a local sync
	source, err := ioutil.TempDir("/tmp", "gosync.source.")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(source)

	destination, err := ioutil.TempDir("/tmp", "gosync.dest.")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(destination)

	makeFiles(testcase.SourceFiles, source)
	makeFiles(testcase.DestFiles, destination)

	opts := &Options{
		SourceHost:    "localhost",
		SourceUDPPort: 30000,

		DestinationHost:    "localhost",
		DestinationUDPPort: 30001,

		Path:        source,
		Destination: destination,

		FollowLinks: false,
		BlockSize:   testcase.BlockSize,
	}

	listenerDone := make(chan bool)

	var outstats *TransferStats
	// gorouting to handle source side
	go func() {
		defer close(listenerDone)

		ln, err := net.Listen("tcp", "localhost:4038")

		if err != nil {
			t.Error(err)
			return
		}
		conn, err := ln.Accept()

		if err != nil {
			t.Error(err)
			return
		}

		// clean up ln
		defer func() {
			lnFile, err := ln.(*net.TCPListener).File()
			if err != nil {
				t.Error(err)
			}

			if err := ln.Close(); err != nil {
				t.Error(err)
			}

			if err := lnFile.Close(); err != nil {
				t.Error(err)
			}
		}()

		outstats, err = SyncOutgoing(conn, opts)

		if err != nil {
			t.Error(err)
			return
		}

	}()

	conn, err := net.Dial("tcp", "localhost:4038")
	if err != nil {
		t.Error(err)
		return nil
	}

	stats, err := SyncIncoming(conn, opts)

	if err != nil {
		t.Error(err)
	}

	<-listenerDone

	if stats.NetStats.ResentDestinationPackets != outstats.NetStats.ResentDestinationPackets {
		t.Error(fmt.Sprintf("stats and outstats ResentDestinationPackets "+
			"should be equal (%v != %v)",
			stats.NetStats.ResentDestinationPackets,
			outstats.NetStats.ResentDestinationPackets))
	}
	if stats.NetStats.ResentSourcePackets != outstats.NetStats.ResentSourcePackets {
		t.Error(fmt.Sprintf("stats and outstats ResentSourcePackets "+
			"should be equal (%v != %v)",
			stats.NetStats.ResentSourcePackets,
			outstats.NetStats.ResentSourcePackets))
	}

	return stats

}
