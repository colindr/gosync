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

// Packeter manages incoming and outgoing packets
// It keeps a copy of all packets sent until it's confirmed
// that they have been received.
// It also gathers incoming packets until all content groups
// are recieved so the can be decoded.
type Packeter struct {
	sendCache       map[int64]Packet
	receiveCache    map[int64]Packet

	PacketChannel   chan Packet

	LastPacketSent        int64
	LastPacketReceived    int64
	LastPacketDecoded     int64
}

// PacketerStatus is part of the status that is sent back and
// forth by the TCPer. It's source/destination agnostic so the
// Packeter can be used identically on both sides.

type PacketerStatus struct {
	LastPacketReceived int64
	ResendPackets      []int64
	LastPacketSent     int64
}

func NewPacketer () *Packeter {
	return &Packeter{}
}

// SendPackets inserts the supplied packets into the sendCache,
// adds them to the PacketChannel, increments LastPacketSent and
// returns the number of the last packet sent
func (packeter *Packeter) SendPackets(packets []Packet) (int64, error) {
    // insert into sendCache

    // increment packeter.LastPacketSent

    return packeter.LastPacketSent, nil
}

// ReceivePacket inserts the packet into the receiveCache, which
// the Decoder goroutine is constantly iterating over and decoding.
// This function also optionally updates the LastPacketReceived.
func (packeter *Packeter) ReceievePacket(packet Packet) error {
	// insert into receiveCache

	// increment packeter.LastPacketReceived
	return nil
}

// ReceivePacketerStatusUpdate is called by a manger, it informs this
// packeter of the status of it's counterpart packeter. With this new
// information this packeter must:
//   - delete unneeded entries from the sendCache
//   - resend any packets that the other packeter thinks needs resending
//   - determine what packets the other packeter needs to resend
//   - respond with this packeter's status, including resend list
// TODO: should we include some timing information with the status update
//       so that we can better determine whether or not it's appropriate
//       to request resent packets?
func (packeter *Packeter) ReceivePacketerStatusUpdate(status PacketerStatus) PacketerStatus{
	// Delete any packets that were successfully sent
	packeter.deleteSentPackets(status.LastPacketReceived)
	// Resend any un-received packets
	packeter.resendPackets(status.ResendPackets)

	// Request any un-received packets
	return PacketerStatus{
		LastPacketReceived: packeter.LastPacketReceived,
		LastPacketSent: packeter.LastPacketSent,
		ResendPackets: packeter.determineResendPackets(status.LastPacketSent),
	}
}


func (packeter *Packeter) deleteSentPackets(lastReceived int64){

}

func (packeter *Packeter) resendPackets(packetNumbers []int64){

}

func (packeter *Packeter) determineResendPackets(lastSent int64) []int64 {
	return make([]int64, 1)
}
