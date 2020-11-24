package utils

import (
	"fmt"
	"testing"
)

func TestSliceDeduplication(t *testing.T) {
	s := SliceDeduplication([]string{"7d0ad09a-11ab-4147-9734-974276f397d1"})
	fmt.Println(s)
}

func TestIntersectStrSlice(t *testing.T) {
	s1 := []string{"2", "2"}
	s2 := []string{"1", "22", "3", "4"}
	result := IntersectAndDeduplicateStrSlice(s1, s2)
	t.Log(result)
}
