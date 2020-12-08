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
