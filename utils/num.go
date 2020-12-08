package utils

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"strconv"
	"strings"
)

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

func ParseInt(ctx context.Context, num string) int {
	d, err := strconv.Atoi(num)
	if err != nil{
		log.Warn(ctx, "parse number failed", log.Err(err), log.String("num", num))
		return 0
	}
	return d
}

func ParseInt64(ctx context.Context, num string) int64 {
	d, err := strconv.ParseInt(num, 10, 64)
	if err != nil{
		log.Warn(ctx, "parse number failed", log.Err(err), log.String("num", num))
		return 0
	}
	return d
}
func StringToStringArray(ctx context.Context, str string) []string {
	if str == "" {
		return nil
	}
	return strings.Split(str, ",")
}