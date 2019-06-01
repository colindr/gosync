package transfer

import (
	"golang.org/x/crypto/blake2b"
	"hash"
)

func Signature(data []byte) (hash.Hash, error) {
	var key []byte
	key = nil
	h, err := blake2b.New256(key)
	h.Write(data)
	return h, err
}
