package transfer

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"time"
)

func UDPSender(host string, port int, opts *Options, manager Manager) {

	gob.Register(&Packet{})

	// make udp conn
	var raddr *net.UDPAddr
	raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%d", host, port))
	if err != nil {
		manager.ReportError(err)
		return
	}
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		manager.ReportError(err)
		return
	}

	// close it at the end of this func
	defer conn.Close()
	// tell the packeter that we're done
	defer manager.Packeter().SenderDone()

	var buf bytes.Buffer

	for packet := range manager.Packeter().PacketChannel {
		encoder := gob.NewEncoder(&buf)

		if manager.Done() || manager.Error() != nil {
			break
		}

		if err := encoder.Encode(packet); err != nil {
			manager.ReportError(err)
			return
		}

		var n int
		if n, err = conn.WriteTo(buf.Bytes(), raddr); err != nil {
			manager.ReportError(err)
			return
		}

		if n != buf.Len() {
			manager.ReportError(fmt.Errorf("didn't send full packet"))
		}

		buf.Reset()
		Debug(fmt.Sprintf("Sent Packet %v", packet))
	}

	Debug("UDP Sender Done")
}

// Helper func for debugging gob encoders/decoders.
// What I found is that encoders/decoders need to be paired up.
// Essentially the first time an encoder encodes data, it includes
// more information so that the decoder knows how to decode the data.
// The second time, if you re-use the encoder, it will send less data
// assuming the decoder already decoded the first piece of data and now
// knows how to decode the same thing in the future without instructions.
// This is great because it's efficient but wasn't super clear. So
// if we re-use the encoder we must re-use the same decoder.
// We also must decode in the same order as we encode.  In my implementation
// I make sure that we're always encoding and decoding in the same order,
// and thus we use the same encoder/decoder for each udp send/receive pair.
// The packet contents themselves also are encoded using a single encoder,
// and decoded also using the same decoder.
func verifyEncodeDecode(packet Packet) {

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(packet); err != nil {
		panic(err)
	}

	newp := &Packet{}
	decoder := gob.NewDecoder(&buf)
	if err := decoder.Decode(newp); err != nil {
		panic(err)
	}
}

func UDPReceiver(host string, port int, opts *Options, manager Manager) {
	// listen to incoming udp packets
	conn, err := net.ListenPacket("udp", fmt.Sprintf("%v:%d", host, port))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	// tell the packeter that receiving is done
	defer manager.Packeter().ReceiverDone()

	gob.Register(&Packet{})

	var reader bytes.Buffer

	for !manager.Done() && manager.Error() == nil {
		decoder := gob.NewDecoder(&reader)

		t := time.Now().Add(time.Duration(100) * time.Millisecond)
		if err := conn.SetReadDeadline(t); err != nil {
			manager.ReportError(err)
			return
		}

		var n int
		buf := make([]byte, 4096)
		n, _, err = conn.ReadFrom(buf)
		if err != nil {
			neterr, ok := err.(net.Error)
			if !ok {
				manager.ReportError(err)
				return
			}

			if neterr.Timeout() {
				Debug("timeout reading")
				continue
			} else {
				manager.ReportError(err)
				return
			}
		}

		if n, err = reader.Write(buf[:n]); err != nil {
			manager.ReportError(err)
			return
		}

		packet := &Packet{}

		if err := decoder.Decode(packet); err != nil {
			manager.ReportError(err)
			return
		}
		Debug(fmt.Sprintf("Got Packet %v", packet))

		manager.Packeter().ReceievePacket(*packet)
	}

	Debug("UDP Receiver Done")

}
