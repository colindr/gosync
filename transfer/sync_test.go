package transfer

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)


type SyncTestCaseFilePiece struct {
	Character     rune
	Num           int
}

type SyncTestCaseFile struct {
	RelPath       string
	Mode          os.FileMode
	Pieces        []SyncTestCaseFilePiece
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


func TestAbsPathVerify(t *testing.T) {
	req := &Request{
		RequestID: uuid.New(),

		Type: Local,

		Path: "a",
		Destination: "/b",

		FollowLinks: false,
		BlockSize: 10,
	}

	_, err := SyncLocal(req)

	if err == nil {
		t.Error("Should have gotten an absolute path error")
	}

	req.Path = "/a"
	req.Destination = "b"

	_, err = SyncLocal(req)

	if err == nil {
		t.Error("Should have gotten an absolute path error")
	}

	req.Path = "a"
	req.Destination = "b"

	_, err = SyncLocal(req)

	if err == nil {
		t.Error("Should have gotten an absolute path error")
	}
}

func TestSimpleSyncLocal(t *testing.T) {

	testcase := SyncTestCase{
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
		DestFiles: []SyncTestCaseFile{},
		BlockSize: 10,
		BytesSent: 30,
		BytesSame: 0,
		Directories: 1,
		Files: 2,
	}

	buildAndRunLocalSyncTest(t, testcase)


	// run the same test with a block size larger than the file size
	testcase.BlockSize = 4096
	buildAndRunLocalSyncTest(t, testcase)

}

func TestChecksumSyncLocal(t *testing.T) {

	testcase := SyncTestCase{
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
		BlockSize: 10,
		BytesSent: 10,
		BytesSame: 10,
		Directories: 1,
		Files: 1,
	}

	buildAndRunLocalSyncTest(t, testcase)

	// run the same test with a block size larger than the file size
	testcase.BlockSize = 100
	buildAndRunLocalSyncTest(t, testcase)
}


func assertFiles(t *testing.T, stats *TransferStats, files []SyncTestCaseFile, dir string){

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


func makeFiles(files []SyncTestCaseFile, dir string){
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
		if err != nil { panic(err) }
	}
}

// buildAndRunLocalSyncTest is a helper function that
func buildAndRunLocalSyncTest(t *testing.T, testcase SyncTestCase) *TransferStats{

	source, err := ioutil.TempDir("/tmp", "gosync.source.")
	if err != nil { panic(err) }
	defer os.RemoveAll(source)

	destination, err := ioutil.TempDir("/tmp", "gosync.dest.")
	if err != nil { panic(err) }
	defer os.RemoveAll(destination)

	makeFiles(testcase.SourceFiles, source)
	makeFiles(testcase.DestFiles, destination)

	req := &Request{
		RequestID: uuid.New(),

		Type: Local,

		Path: source,
		Destination: destination,

		FollowLinks: false,
		BlockSize: testcase.BlockSize,
	}

	stats, err := SyncLocal(req)

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