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

type temp struct {
	Name string
}

func TestTemp2(t *testing.T) {
	strs := []*temp{&temp{Name: "sdf"}}
	strs = append(strs, &temp{Name: "2222"})
	for _, item := range strs {
		t.Log(item.Name)
	}
}
