package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
)

func Hash(obj interface{}) (string, error) {
	hash, err := HashBytes(obj)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash), nil
}

func HashBytes(obj interface{}) ([]byte, error) {
	// sha1 is enough
	h := sha1.New()
	err := json.NewEncoder(h).Encode(obj)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
