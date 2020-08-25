package utils

import (
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"strconv"
)

func GetPager(pageStr string, pageSizeStr string) dbo.Pager {
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = constant.DefaultPageIndex
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		pageSize = constant.DefaultPageSize
	}
	if page == 0 || pageSize == 0 {
		return dbo.NoPager
	}
	return dbo.Pager{Page: page, PageSize: pageSize}
}

// Obsolete
//type Pager struct {
//	PageIndex int64
//	PageSize  int64
//}
//
//func GetPager(pageIndexStr string, pageSizeStr string) Pager {
//	pageIndex, err := strconv.ParseInt(pageIndexStr, 10, 64)
//	if err != nil {
//		pageIndex = constant.DefaultPageIndex
//	}
//	pageSize, err := strconv.ParseInt(pageSizeStr, 10, 64)
//	if err != nil {
//		pageSize = constant.DefaultPageSize
//	}
//	if pageIndex == 0 || pageSize == 0 {
//		return Pager{PageIndex: 0, PageSize: 0}
//	}
//	return Pager{PageIndex: pageIndex, PageSize: pageSize}
//}
