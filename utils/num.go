package utils

func NumToNumArray(num int) []int{
	ret := make([]int, 0)
	index := 1
	for curNum := num; curNum > 0; curNum = curNum >> 1 {
		if curNum & 0x01 != 0 {
			ret = append(ret, index)
		}
		index ++
	}
	return ret
}
