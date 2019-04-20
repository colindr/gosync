package transfer

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"path"
)

// Direction - a Request is either for a pull or a push
type Direction uint8

// Local means requester will read and write data
const Local Direction = 0

// Incoming means the requester wants to read data
const Incoming Direction = 1

// Outgoing means the requester wants to write data
const Outgoing Direction = 2

// default block length
const DefaultBlockLength int = 2048

// Request - information to initiate a transfer request
type Request struct {
	RequestID   uuid.UUID

	RequesterHost        string
	RequesterUDPPort     int

	Host        string
	Port        int

	Direction   Direction

	Path        string
	Destination string

	FollowLinks bool
	BlockSize   int
}

// Once a transfer is requested and responded to, the relevant
// information is copied into Options.  This options contains the
// request options like Path, Destination, FollowLinks, and BlockSize,
// as well as host/port options.  There is no direction on the Options
// object since it's the same object at the source and destination.
type Options struct {
	Path               string
	Destination        string

	FollowLinks        bool
	BlockSize          int

	SourceHost         string
	SourceUDPPort      int

	DestinationHost    string
	DestinationUDPPort int

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
func (opts Options) Verify() error {

	if ! path.IsAbs(opts.Path){
		return errors.New(fmt.Sprintf(
			"Path attribute is not an absolute path: %v", opts.Path))
	}

	if ! path.IsAbs(opts.Destination){
		return errors.New(fmt.Sprintf(
			"Destination attribute is not an absolute path: %v", opts.Destination))
	}

	return nil
}