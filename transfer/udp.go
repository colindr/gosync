package transfer

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
)

func UDPSender(host string, port int, opts *Options, packeter *Packeter, manager Manager){

	// make udp conn
	var laddr,raddr *net.UDPAddr
	laddr = nil
	raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%d", host, port))
	if err != nil {
		manager.ReportError(err)
		return
	}
	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		manager.ReportError(err)
		return
	}

	// close it at the end of this func
	defer conn.Close()

	encoder := gob.NewEncoder(conn)

	for packet := range packeter.PacketChannel {

		if manager.Done() || manager.Error()!= nil{
			break
		}

		if err := encoder.Encode(packet); err != nil {
			manager.ReportError(err)
			return
		}
	}

}


func UDPReceiver(host string, port int, opts *Options, packeter *Packeter, manager Manager){
	// listen to incoming udp packets
	conn, err := net.ListenPacket("udp", fmt.Sprintf("%v:%d", host, port))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	var packet Packet
	bytearray := make([]byte, 4096)
	buff := bytes.NewBuffer(bytearray)

	decoder := gob.NewDecoder(buff)

	for (!manager.Done() && manager.Error()==nil){

		if _, _, err := conn.ReadFrom(bytearray); err != nil {
			manager.ReportError(err)
			return
		}

		if err := decoder.Decode(packet); err != nil {
			manager.ReportError(err)
			return
		}

		packeter.ReceievePacket(packet)
	}


}