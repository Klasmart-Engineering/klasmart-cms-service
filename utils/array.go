package utils

func CheckInStringArray(str string, arr []string) bool{
	for i := range arr {
		if str == arr[i] {
			return true
		}
	}
	return false
}
