package utils

import (
	"testing"
)

func TestHashBytes(t *testing.T) {
	request := &struct {
		AAA string
		BBB int
	}{
		AAA: "aaa",
		BBB: 123,
	}

	hash := HashBytes(request)
	if hash == nil {
		t.Error("caculate hash failed")
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

	hash := Hash(request)
	if hash == "" {
		t.Error("caculate hash failed")
	}
}
