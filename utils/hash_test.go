package utils

import (
	"bytes"
	"database/sql"
	"testing"
)

func TestHashBytes(t *testing.T) {
	type ss struct {
		AAA string
		BBB int
		CCC sql.NullString
	}
	request1 := &ss{
		AAA: "aaa",
		BBB: 123,
		CCC: sql.NullString{String: "/", Valid: false},
	}

	hash1, err := HashBytes(request1)
	if err != nil {
		t.Errorf("caculate hash failed due to %v", err)
	}

	request2 := &ss{
		AAA: "aaa",
		BBB: 123,
		CCC: sql.NullString{String: "/aaaa", Valid: true},
	}

	hash2, err := HashBytes(request2)
	if err != nil {
		t.Error("caculate hash failed")
	}

	if bytes.Equal(hash1, hash2) {
		t.Error("hash must not be equal")
	}

}

func BenchmarkHashBytes(b *testing.B) {
	request := &struct {
		AAA string
		BBB int
	}{
		AAA: "aaa",
		BBB: 123,
	}

	for n := 0; n < b.N; n++ {
		HashBytes(request)
	}
}

func TestHash(t *testing.T) {
	request := &struct {
		AAA string
		BBB int
	}{
		AAA: "aaa",
		BBB: 123,
	}

	_, err := Hash(request)
	if err != nil {
		t.Errorf("caculate hash failed due to %v", err)
	}
}
