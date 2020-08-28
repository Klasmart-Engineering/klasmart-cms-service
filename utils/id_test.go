package utils

import (
	"fmt"
	"testing"
)

func TestNewId(t *testing.T) {
	t.Log(NewID())
	t.Log(NewID())
	t.Log(len(NewID()))
	t.Log(NewID())
}

func TestUniqueIdList(t *testing.T) {
	r := SliceDeduplication([]string{"a", "b", "2", "c", "a", "2"})
	fmt.Println(r)
}
