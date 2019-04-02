package transfer

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"os"
	"path"
)

// Direction - a Request is either for a pull or a push
type Type uint8

// Local means requester will read and write data
const Local Type = 0

// Incoming means the requester wants to read data
const Incoming Type = 1

// Outgoing means the requester wants to write data
const Outgoing Type = 2

// default block length
const DefaultBlockLength int = 2048

// Request - information to initiate a transfer request
type Request struct {
	RequestID   uuid.UUID

	Host        string
	Port        int

	Type        Type

	Path        string
	Destination string

	FollowLinks bool
	BlockSize   int
}


type FileInfo struct {
	FileInfo        os.FileInfo
	SourcePath      string
	DestinationPath string
}

// RequestResponse - response to a TransferRequest
type RequestResponse struct {
	Accepted  bool
	Reason    string
	RequestID uuid.UUID
	UDPPort   int
}

// Verify will return an error if there's anything
// wrong with the request.  Currently only checks that
// Path and Destination are absolute.
func (req Request) Verify() error {

	if ! path.IsAbs(req.Path){
		return errors.New(fmt.Sprintf(
			"Path attribute is not an absolute path: %v", req.Path))
	}

	if ! path.IsAbs(req.Destination){
		return errors.New(fmt.Sprintf(
			"Destination attribute is not an absolute path: %v", req.Destination))
	}

	return nil
}