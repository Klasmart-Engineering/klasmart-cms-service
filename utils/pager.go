package utils

import (
	"strconv"

	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
)

// recommended that GET request use
func GetDboPager(pageStr string, pageSizeStr string) dbo.Pager {
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

// recommended that POST & application/json request use
func GetDboPagerFromInt(page int, pageSize int) dbo.Pager {
	if page == -1 && pageSize == -1 {
		return dbo.NoPager
	}

	if page < 0 || pageSize < 0 {
		page = constant.DefaultPageIndex
		pageSize = constant.DefaultPageSize
	}

	return dbo.Pager{Page: page, PageSize: pageSize}
}

// Obsolete
type Pager struct {
	PageIndex int64
	PageSize  int64
}

//
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
