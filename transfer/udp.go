package transfer

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
)

func UDPSender(host string, port int, opts *Options, manager Manager){

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

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	for packet := range manager.Packeter().PacketChannel {

		if manager.Done() || manager.Error()!= nil{
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
	}
}


func verifyEncodeDecode(packet Packet) {

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(packet); err!= nil{
		panic(err)
	}

	newp := &Packet{}
	decoder := gob.NewDecoder(&buf)
	if err := decoder.Decode(newp); err!= nil{
		panic(err)
	}
}


func UDPReceiver(host string, port int, opts *Options, manager Manager){
	// listen to incoming udp packets
	conn, err := net.ListenPacket("udp", fmt.Sprintf("%v:%d", host, port))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	gob.Register(&Packet{})

	var reader bytes.Buffer
	decoder := gob.NewDecoder(&reader)

	for (!manager.Done() && manager.Error()==nil){

		var n int
		buf := make([]byte, 4096)
		if n, _, err = conn.ReadFrom(buf); err != nil {
			manager.ReportError(err)
			return
		}

		if n == 0 {
			manager.ReportError(fmt.Errorf("Bad"))
		}

		if n, err = reader.Write(buf[:n]); err != nil{
			manager.ReportError(err)
			return
		}


		packet := &Packet{}

		if err := decoder.Decode(packet); err != nil {
			manager.ReportError(err)
			return
		}

		manager.Packeter().ReceievePacket(*packet)
	}


}