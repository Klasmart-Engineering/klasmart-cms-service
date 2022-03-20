package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

func Hash(obj interface{}) string {
	return hex.EncodeToString(HashBytes(obj))
}

func HashBytes(obj interface{}) []byte {
	// sha1 is enough
	h := sha1.New()
	fmt.Fprintf(h, "%v", obj)
	return h.Sum(nil)
}
