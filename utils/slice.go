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

func FilterStrings(source []string, whitelist, blacklist []string) []string {
	var result []string
	for _, item := range source {
		pass := false
		for _, target := range whitelist {
			if item == target {
				pass = true
			}
		}
		if !pass {
			find := false
			for _, target := range blacklist {
				if item == target {
					find = true
				}
			}
			if !find {
				pass = true
			}
		}
		if pass {
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

func IntersectAndDeduplicateStrSlice(slice1 []string, slice2 []string) []string {
	m := make(map[string]int)
	m2 := make(map[string]int)
	for _, v := range slice1 {
		m[v]++
	}
	for _, v := range slice2 {
		times, _ := m[v]
		if times > 0 {
			m2[v]++
		}
	}
	result := make([]string, 0, len(m2))
	for key, _ := range m2 {
		result = append(result, key)
	}
	return result
}
