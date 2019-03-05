package transfer

import "github.com/google/uuid"

// Direction - a Request is either for a pull or a push
type Direction int

// Incoming means the requester wants to read data
const Incoming Direction = 0

// Outgoing means the requester wants to write data
const Outgoing Direction = 0

// default block length
const DefaultBlockLength int = 2048

// Request - information to initiate a transfer request
type Request struct {
	RequestID   uuid.UUID

	Host        string
	Port        int

	Direction   Direction

	Path        string
	Destination string

	FollowLinks bool
}

// RequestResponse - response to a TransferRequest
type RequestResponse struct {
	Accepted  bool
	Reason    string
	RequestID uuid.UUID
	UDPPort   int
}
