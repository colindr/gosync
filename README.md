# gosync
Experiment in golang to sync files. 

## Goals
1. Sync a file locally from one path to another
2. Sync should send minimal diff
3. Sync over tcp between client and daemon
4. Sync over udp between client and daemon

## Stretch Goals
1. TLS support
2. Bandwidth Control

## Current TODO
- [x] Add Statistics Gathering
- [ ] Handle Symlinks
- [x] Make Paths Absolute
- [x] Parameterize SyncLocal Tests
- [ ] Add Signature Hash
- [x] Start net communication implementation
- [ ] Finish v1 of net communication
  - [x] Diagram design
  - [ ] Implement Packet Encoding
  - [x] Implement SourceManager
  - [ ] Implement DestinationManager
  - [ ] Implement Packeter SendPackets/ReceivePacket
  - [ ] Implement Packeter resendPackets/deleteSentPackets/determineResendPackets
  - [ ] Implement Decoder
  - [ ] Implement TCP Loop goroutine
  - [ ] Implement UDP Sender/Receiver goroutines
- [ ] Test v1 of net communication
- [ ] Support preserving file mode
- [ ] Add NoOp Signature for same mtime/size
