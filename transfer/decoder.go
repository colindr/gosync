package transfer

import (
	"bytes"
	"encoding/gob"
)

func DecodePackets(manager Manager, packeter *Packeter){
	// loop until we're done or an error is reported
	for (manager.Done() || manager.Error()!=nil) {

		// iterate from LastPacketDecoded to LastPacketReceived
		// if we get consecutive packets up to an end packet then
		// we put all the content into a buffer and decode it
		start := packeter.LastPacketDecoded + 1
		i := start
		end := packeter.LastPacketReceived

		for i<end {

			packet, ok := packeter.receiveCache[i]
			if !ok{
				// There's a hole in our receiveCache dear 'liza,
				// break and try again later
				break

			} else if packet.IsEndPacket {
				// We've found a consecutive set of packets!
				// ReadDecodeAndSend them
				err := readDecodeAndSend(manager, packeter, start, i)
				if err!= nil{
					manager.ReportError(err)
				}

				// The start of the next packetset is i + 1
				start = i + 1
			}
			i++
		}

	}
}

func readDecodeAndSend(manager Manager, packeter *Packeter, start uint64, end uint64) error{
	var buff bytes.Buffer

	// read the content from our packet range in the buffer
	// for decoding
	for i:=start; i<=end; i++ {
		if _, err:=buff.Read(packeter.receiveCache[i].Content); err!=nil{
			return err
		}
		// remove from the receiveCache
		delete(packeter.receiveCache, i)
	}

	// update the LastPacketDecoded
	packeter.LastPacketDecoded = end

	decoder := gob.NewDecoder(&buff)

	switch packeter.receiveCache[start].ContentType {

	case FileInfoPacket:
		var fi FileInfo
		if err := decoder.Decode(&fi); err != nil{
			return err
		}
		manager.QueueFileInfo(fi)
	case SignaturePacket:
		var sig Checksum
		if err := decoder.Decode(&sig); err != nil{
			return err
		}
		manager.QueueSignature(sig)
	case DeltaPacket:
		var delta Delta
		if err := decoder.Decode(&delta); err != nil{
			return err
		}
		manager.QueueDelta(delta)
	}
	return nil
}