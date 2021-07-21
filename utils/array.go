package utils

import (
	"context"
	"fmt"
	"strings"
)

func CheckInStringArray(str string, arr []string) bool {
	for i := range arr {
		if str == arr[i] {
			return true
		}
	}
	return false
}

func StringCountRange(ctx context.Context, prefix string, suffix string, count int) string {
	items := make([]string, count)
	for i := 0; i < count; i++ {
		items[i] = fmt.Sprintf("%s%d%s", prefix, i, suffix)
	}
	return strings.Join(items, " ")
}
