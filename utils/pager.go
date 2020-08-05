package utils

import (
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"strconv"
)

type Pager struct {
	PageIndex     int64
	PageSize int64
}
func GetPager(pageIndexStr string, pageSizeStr string) Pager {
	pageIndex, err := strconv.ParseInt(pageIndexStr, 10, 64)
	if err != nil {
		pageIndex = constant.DefaultPageIndex
	}
	pageSize, err := strconv.ParseInt(pageSizeStr, 10, 64)
	if err != nil {
		pageSize = constant.DefaultPageSize
	}
	if pageIndex == 0 || pageSize == 0 {
		return Pager{PageIndex: 0, PageSize: 0}
	}
	return Pager{PageIndex: pageIndex, PageSize: pageSize}
}
