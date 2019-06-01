package transfer

import (
	"bytes"
	"encoding/gob"
	"time"
)

func DecodePackets(manager Manager) {
	// loop until we're done or an error is reported
	for !manager.Done() && manager.Error() == nil {
		// iterate from LastPacketDecoded to LastPacketReceived
		// if we get consecutive packets up to an end packet then
		// we put all the content into a buffer and decode it
		start := manager.Packeter().LastPacketDecoded + 1
		i := start
		end := manager.Packeter().LastPacketReceived

		for i <= end {

			manager.Packeter().receiveCacheMutex.RLock()
			packet, ok := manager.Packeter().receiveCache[i]
			manager.Packeter().receiveCacheMutex.RUnlock()

			if !ok {
				// There's a hole in our receiveCache dear 'liza,
				// break and try again later
				break

			} else if packet.IsEndPacket {
				// We've found a consecutive set of packets!
				// ReadDecodeAndSend them
				err := readDecodeAndSend(manager, start, i)
				if err != nil {
					manager.ReportError(err)
				}

				// The start of the next packetset is i + 1
				start = i + 1
			}
			i++
		}
		// NOTE: if this sleep isn't here then the whole process can get
		// stuck, seemingly because this goroutine hogs all available
		// CPU or something dumb like that?
		time.Sleep(time.Millisecond * 100)

	}
}

func readDecodeAndSend(manager Manager, start uint64, end uint64) error {
	var buff bytes.Buffer

	var contentType PacketContentType
	// read the content from our packet range in the buffer
	// for decoding
	for i := start; i <= end; i++ {
		manager.Packeter().receiveCacheMutex.RLock()
		// this function assumes it's the only thing that removes stuff
		// from the receive cache, and that all packets from start to
		// end exist in the cache, so we don't check "ok"
		packet, _ := manager.Packeter().receiveCache[i]
		manager.Packeter().receiveCacheMutex.RUnlock()

		if _, err := buff.Write(packet.Content); err != nil {
			return err
		}

		contentType = packet.ContentType

		// remove from the receiveCache
		manager.Packeter().receiveCacheMutex.Lock()
		delete(manager.Packeter().receiveCache, i)
		manager.Packeter().receiveCacheMutex.Unlock()
	}

	decoder := gob.NewDecoder(&buff)

	switch contentType {

	case FileInfoPacket:
		var fi FileInfo
		if err := decoder.Decode(&fi); err != nil {
			return err
		}
		manager.QueueFileInfo(fi)
	case SignaturePacket:
		var sig Checksum
		if err := decoder.Decode(&sig); err != nil {
			return err
		}
		manager.QueueSignature(sig)
	case DeltaPacket:
		var delta Delta
		if err := decoder.Decode(&delta); err != nil {
			return err
		}
		manager.QueueDelta(delta)
	}

	// update the LastPacketDecoded
	manager.Packeter().LastPacketDecoded = end

	return nil
}
