package da

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
)

type IReportDA interface {
	DataAccessor
	ITeacherLoadAssessment
	ITeacherLoadLesson
	IStudentProgressAssignment
	IStudentProgressLearnOutcomeAchievement
	IClassAttendance
	ILearnOutcome
	ILearnerWeekly
}
type ReportDA struct {
	BaseDA
	learnerReportOverviewCache *LazyRefreshCache
}

var _reportDA *ReportDA
var _reportDAOnce sync.Once

func GetReportDA() IReportDA {
	_reportDAOnce.Do(func() {
		_reportDA = new(ReportDA)

		learnerReportOverviewCache, err := NewLazyRefreshCache(&LazyRefreshCacheOption{
			RedisKeyPrefix:  RedisKeyPrefixReportLearnerReportOverview,
			Expiration:      constant.ReportQueryCacheExpiration,
			RefreshDuration: constant.ReportQueryCacheRefreshDuration,
			RawQuery:        _reportDA.getLearnerReportOverview})
		if err != nil {
			log.Panic(context.Background(), "create learner report overview cache failed", log.Err(err))
		}

		_reportDA.learnerReportOverviewCache = learnerReportOverviewCache

	})
	return _reportDA
}
