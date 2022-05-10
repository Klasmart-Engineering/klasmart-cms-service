package constant

import "time"

const (
	MinCacheExpire = time.Second * 30
	MaxCacheExpire = time.Minute * 10
)

const (
	ContentFolderQueryCacheRefreshDuration = time.Minute
	ContentFolderQueryCacheExpiration      = 0 // never expire

	SharedContentQueryCacheRefreshDuration = time.Minute
	SharedContentQueryCacheExpiration      = time.Minute * 10 // never expire
)

const (
	ReportQueryCacheExpiration      = time.Minute * 5
	ReportQueryCacheRefreshDuration = time.Minute
)
