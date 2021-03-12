package utils

import "strings"

type PredicateBuilder struct {
	formats []string
	values  []interface{}
}

func NewPredicateBuilder() *PredicateBuilder {
	return &PredicateBuilder{}
}

func FalsePredicateBuilder() *PredicateBuilder {
	return NewPredicateBuilder().Append("1 = 0")
}

func (pb *PredicateBuilder) Append(format string, values ...interface{}) *PredicateBuilder {
	pb.formats = append(pb.formats, format)
	pb.values = append(pb.values, values...)
	return pb
}

func (pb *PredicateBuilder) Join(sep string, left string, right string) *PredicateBuilder {
	return NewPredicateBuilder().Append(left+strings.Join(pb.formats, sep)+right, pb.values...)
}

func (pb *PredicateBuilder) Or() *PredicateBuilder {
	return pb.Join(" or ", "(", ")")
}

func (pb *PredicateBuilder) And() *PredicateBuilder {
	return pb.Join(" and ", "(", ")")
}

func (pb *PredicateBuilder) Merge(other *PredicateBuilder) *PredicateBuilder {
	if other == nil {
		return pb
	}
	pb.formats = append(pb.formats, other.formats...)
	pb.values = append(pb.values, other.values...)
	return pb
}

func (pb *PredicateBuilder) Raw() ([]string, []interface{}) {
	return pb.formats, pb.values
}
