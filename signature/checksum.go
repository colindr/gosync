package signature

import (
	"golang.org/x/crypto/blake2b"
	"hash"
)

func Signature(data []byte) (hash.Hash, error) {
	return blake2b.New256(data)
}