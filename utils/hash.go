package utils

import (
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
)

func Hash(obj interface{}) string {
	return hex.EncodeToString(HashBytes(obj))
}

func HashBytes(obj interface{}) []byte {
	// sha1 is enough
	h := sha1.New()
	gob.NewEncoder(h).Encode(obj)
	return h.Sum(nil)
}
