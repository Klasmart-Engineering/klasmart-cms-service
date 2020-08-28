package utils

func SliceDeduplication(s []string) []string {
	temp := make(map[string]bool)
	for i := range s {
		temp[s[i]] = true
	}

	result := make([]string, 0)
	for k, v := range temp {
		if v {
			result = append(result, k)
		}
	}
	//sort.Strings(retIdList)
	return result
}
