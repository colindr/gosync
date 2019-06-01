package transfer

import (
	"bytes"
	"sync"
)

type PacketContentType uint8

const PACKET_CHANNEL_SIZE = 100

const FileInfoPacket PacketContentType = 0
const SignaturePacket PacketContentType = 1
const DeltaPacket PacketContentType = 2

type Packet struct {
	PacketID    uint64
	IsEndPacket bool
	ContentType PacketContentType
	Content     []byte
}

const PACKET_CONTENT_LEN = 500

// Packeter manages incoming and outgoing packets
// It keeps a copy of all packets sent until it's confirmed
// that they have been received.
// It also gathers incoming packets until all content groups
// are recieved so the can be decoded.
type Packeter struct {
	sendCache    map[uint64]Packet
	receiveCache map[uint64]Packet

	receiveCacheMutex sync.RWMutex

	packetMutex sync.Mutex

	senderDone   bool
	receiverDone bool

	PacketChannel chan Packet

	LastDeletedPacket  uint64
	LastPacketSent     uint64
	LastPacketReceived uint64
	LastPacketDecoded  uint64
}

// PacketerStatus is part of the status that is sent back and
// forth by the TCPer. It's source/destination agnostic so the
// Packeter can be used identically on both sides.
type PacketerStatus struct {
	LastPacketReceived uint64
	ResendPackets      []uint64
	LastPacketSent     uint64
}

func NewPacketer() *Packeter {
	return &Packeter{
		sendCache:    make(map[uint64]Packet),
		receiveCache: make(map[uint64]Packet),

		receiveCacheMutex: sync.RWMutex{},

		packetMutex: sync.Mutex{},

		PacketChannel: make(chan Packet, PACKET_CHANNEL_SIZE),

		LastDeletedPacket:  0,
		LastPacketSent:     0,
		LastPacketReceived: 0,
		LastPacketDecoded:  0,
	}
}

func MakePackets(buffer *bytes.Buffer, packetType PacketContentType) []Packet {
	packets := make([]Packet, (buffer.Len()/PACKET_CONTENT_LEN)+1)
	i := 0
	for buffer.Len() > PACKET_CONTENT_LEN {
		p := Packet{
			ContentType: packetType,
			Content:     buffer.Next(PACKET_CONTENT_LEN),
			IsEndPacket: false,
		}
		packets[i] = p
		i++
	}

	// make last packet
	p := Packet{
		ContentType: packetType,
		Content:     buffer.Next(PACKET_CONTENT_LEN),
		IsEndPacket: true,
	}
	packets[i] = p

	return packets
}

// SendPackets inserts the supplied packets into the sendCache,
// adds them to the PacketChannel, increments LastPacketSent and
// returns the number of the last packet sent
func (packeter *Packeter) SendPackets(packets []Packet) (uint64, error) {
	packeter.packetMutex.Lock()
	// insert into sendCache and add to PacketChannel
	packet_id := packeter.LastPacketSent
	for _, packet := range packets {
		packet_id += 1
		packet.PacketID = packet_id
		packeter.sendCache[packet.PacketID] = packet
		packeter.PacketChannel <- packet

	}

	// increment packeter.LastPacketSent
	packeter.LastPacketSent = packet_id
	packeter.packetMutex.Unlock()

	return packeter.LastPacketSent, nil
}

// ReceivePacket inserts the packet into the receiveCache, which
// the Decoder goroutine is constantly iterating over and decoding.
// This function also optionally updates the LastPacketReceived.
func (packeter *Packeter) ReceievePacket(packet Packet) {
	// insert into receiveCache
	packeter.receiveCacheMutex.Lock()
	packeter.receiveCache[packet.PacketID] = packet
	packeter.receiveCacheMutex.Unlock()

	// increment packeter.LastPacketReceived
	nextLastPackage := packeter.LastPacketReceived + 1

	var ok bool
	packeter.receiveCacheMutex.RLock()
	_, ok = packeter.receiveCache[nextLastPackage]
	packeter.receiveCacheMutex.RUnlock()
	for ok {
		packeter.LastPacketReceived = nextLastPackage
		nextLastPackage++
		packeter.receiveCacheMutex.RLock()
		_, ok = packeter.receiveCache[nextLastPackage]
		packeter.receiveCacheMutex.RUnlock()
	}
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
func (packeter *Packeter) ReceivePacketerStatusUpdate(status PacketerStatus) PacketerStatus {
	// Delete any packets that were successfully sent
	packeter.deleteSentPackets(status.LastPacketReceived)
	// Resend any un-received packets
	packeter.resendPackets(status.ResendPackets)

	// Request any un-received packets
	return PacketerStatus{
		LastPacketReceived: packeter.LastPacketReceived,
		LastPacketSent:     packeter.LastPacketSent,
		ResendPackets:      packeter.determineResendPackets(status.LastPacketSent),
	}
}

func (packeter *Packeter) deleteSentPackets(lastReceived uint64) {

	// iterate between LastDeletedPacket and lastReceived,
	// deleting packets
	for i := packeter.LastDeletedPacket; i <= lastReceived; i++ {
		delete(packeter.sendCache, i)
	}

	// record LastDeletedPacket
	packeter.LastDeletedPacket = lastReceived

}

func (packeter *Packeter) resendPackets(packetNumbers []uint64) {
	// get packets from the sendCache and add to PacketChannel
	for _, packetID := range packetNumbers {
		packeter.PacketChannel <- packeter.sendCache[packetID]
	}
}

func (packeter *Packeter) determineResendPackets(lastSent uint64) []uint64 {
	var neededPackets []uint64
	// simple solution is to ask for all packets between
	// packeter.LastPacketReceived and lastSent that are also not in
	// the receive cache
	// TODO: should we wait for packets to arrive before asking for a resend
	//       in case the status update arrives before the packets?
	for i := packeter.LastPacketReceived + 1; i <= lastSent; i++ {
		if _, ok := packeter.receiveCache[i]; !ok {
			neededPackets = append(neededPackets, i)
		}
	}

	return neededPackets
}

func (packeter *Packeter) Close() {
	// we just close the packet channel which will ensure that
	// the sender stops eventually
	close(packeter.PacketChannel)
}

func (packeter *Packeter) SenderDone() {
	packeter.senderDone = true
}

func (packeter *Packeter) ReceiverDone() {
	packeter.receiverDone = true
}

func (packeter *Packeter) Done() bool {
	return packeter.senderDone && packeter.receiverDone
}
