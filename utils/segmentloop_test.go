package utils

import (
	"context"
	"testing"
)

func TestSegmentLoop(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	SegmentLoop(context.Background(), len(arr), 4, func(start, end int) error {
		t.Log(arr[start:end])
		return nil
	})
}
