package utils

type Int64 []int64

func (v Int64) Len() int {
	return len(v)
}
func (v Int64) Less(i, j int) bool {
	return v[i] < v[j]
}

func (v Int64) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}
