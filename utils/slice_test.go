package utils

import (
	"fmt"
	"testing"
)

func TestSliceDeduplication(t *testing.T) {
	s := SliceDeduplication([]string{"7d0ad09a-11ab-4147-9734-974276f397d1"})
	fmt.Println(s)
}
