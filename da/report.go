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
	ILearnOutcomeReport
	ILearnerWeekly
}
type ReportDA struct {
	BaseDA
	learnerReportOverviewCache   *LazyRefreshCache
	learningOutcomeOverviewCache *LazyRefreshCache
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

		learningOutcomeOverviewCache, err := NewLazyRefreshCache(&LazyRefreshCacheOption{
			RedisKeyPrefix:  RedisKeyPrefixReportLearningOutcomeOverview,
			Expiration:      constant.ReportQueryCacheExpiration,
			RefreshDuration: constant.ReportQueryCacheRefreshDuration,
			RawQuery:        _reportDA.getLearningOutcomeOverview})
		if err != nil {
			log.Panic(context.Background(), "create learning outcome overview cache failed", log.Err(err))
		}

		_reportDA.learningOutcomeOverviewCache = learningOutcomeOverviewCache

	})
	return _reportDA
}
