package transfer

type PacketContentType uint8

const FileInfoPacket PacketContentType = 0
const SignaturePacket PacketContentType = 1
const DeltaPacket PacketContentType = 2

type PacketType uint8

const StartPacket  PacketType = 0
const MiddlePacket PacketType = 1
const EndPacket    PacketType = 2

type Packet struct {
	PacketID     uint
	ContentID    uint
	Type         PacketType
	ContentType  PacketContentType
	Content      []byte
}

// TransferStatus is a struct that represents
// the status of the network communication for a transfer.
// It's the only kind of TCP message that is sent between sides.
type TransferStatus struct {
	SentAllFileInfos          bool
	SentAllSignatures         bool
	SentAllDeltas             bool

	LastPacketSent            int
	LastPacketReceived        int
	ResendPackets             []int

	Failed                    string
}

type OutgoingTransferManager struct {
	Status                    *TransferStatus
	PacketQueue               []Packet
}
