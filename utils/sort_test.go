package utils

import (
	"sort"
	"testing"
)

func TestInt64_Sort(t *testing.T) {
	testData := make(Int64, 0)
	testData = append(testData, 2)
	testData = append(testData, 3)
	testData = append(testData, 1)
	testData = append(testData, 7)
	testData = append(testData, 2)
	sort.Sort(testData)
	t.Log(testData)
}
