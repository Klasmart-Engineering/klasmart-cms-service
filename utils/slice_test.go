package utils

import (
	"fmt"
	"reflect"
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

func TestFilterStrings(t *testing.T) {
	type args struct {
		source    []string
		whitelist []string
		blacklist []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "t1",
			args: args{
				source:    []string{"a", "b", "c", "d"},
				whitelist: []string{"b", "c"},
				blacklist: []string{"c", "d"},
			},
			want: []string{"a", "b", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterStrings(tt.args.source, tt.args.whitelist, tt.args.blacklist); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSliceDeduplicationMap(t *testing.T) {
	origin := []string{"1", "2", "3", "4", "1", "2", "4", "3", "3", "3", "2", "1", "5", "3"}
	wantSlice := []string{"1", "2", "3", "4", "5"}
	wantMap := map[int]int{
		0: 0, 4: 0, 11: 0,
		1: 1, 5: 1, 10: 1,
		2: 2, 7: 2, 8: 2, 9: 2, 13: 2,
		3: 3, 6: 3,
		12: 4,
	}

	resultSlice, resultMap := SliceDeduplicationMap(origin)
	if !reflect.DeepEqual(resultSlice, wantSlice) {
		t.Errorf("SliceDeduplicationMap() got = %+v, want %+v", resultSlice, wantSlice)
	}

	if !reflect.DeepEqual(resultMap, wantMap) {
		t.Errorf("SliceDeduplicationMap() got = %+v, want %+v", resultMap, wantMap)
	}
}

func TestSliceDeduplicationExcludeEmpty(t *testing.T) {
	testData := []string{""}
	result := SliceDeduplicationExcludeEmpty(testData)
	t.Log(len(result))
}

func TestStableSliceDeduplication(t *testing.T) {
	//testData := []string{"1", "2", "1", "3"}
	testData := []string{}
	result := StableSliceDeduplication(testData)
	fmt.Println(result)
}
