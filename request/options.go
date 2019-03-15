package request

import (
	"github.com/google/uuid"
	"os"
)

// Direction - a Request is either for a pull or a push
type Type int

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
