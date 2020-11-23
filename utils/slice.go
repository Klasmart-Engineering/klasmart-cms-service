package utils

func SliceDeduplication(s []string) []string {
	temp := make(map[string]bool)
	for i := range s {
		temp[s[i]] = true
	}

	result := make([]string, 0, len(temp))
	for k, v := range temp {
		if v {
			result = append(result, k)
		}
	}

	return result
}

func ExcludeStrings(source []string, targets []string) []string {
	var result []string
	for _, item := range source {
		find := false
		for _, target := range targets {
			if item == target {
				find = true
			}
		}
		if !find {
			result = append(result, item)
		}
	}
	return result
}

func ContainsStr(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
func ContainsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
