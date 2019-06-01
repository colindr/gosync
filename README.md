# gosync
Experiment in golang to sync files. 

## Goals
1. Sync a file locally from one path to another
2. Sync should send minimal diff
3. Sync over tcp/udp between client and daemon

## Stretch Goals
1. TLS support
2. Bandwidth Control

## Current TODO
- [x] Add Statistics Gathering
- [x] Make Paths Absolute
- [x] Parameterize SyncLocal Tests
- [x] Start net communication implementation
- [x] Finish/Test v1 of net communication
- [x] Fix initial net communication bugs
- [ ] Add net communication stats
- [ ] Implement new udp encoding (can't re-use gob 
  encoder/decoder because packets can get dropped)
- [ ] Implement better packet resend logic
- [ ] Handle Symlinks
- [ ] Support preserving file mode/uid/gid/modtime
- [ ] Add NoOp Signature for same mtime/size
- [ ] Add Signature Hash
- [ ] Make integration tests