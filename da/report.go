package da

import (
	"context"
	"sync"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

type IReportDA interface {
	DataAccessor
	ITeacherLoadAssessment
	ITeacherLoadLesson
	IStudentProgressAssignment
	IStudentProgressLearnOutcomeAchievement
	IClassAttendance
	ILearningOutcomeReport
	ILearnerWeekly
	ISkillCoverage
}
type ReportDA struct {
	BaseDA
	learnerReportOverviewCache   *utils.LazyRefreshCache
	learningOutcomeOverviewCache *utils.LazyRefreshCache
	teacherUsageOverviewCache    *utils.LazyRefreshCache
	skillCoverageCache           *utils.LazyRefreshCache
}

var _reportDA *ReportDA
var _reportDAOnce sync.Once

func GetReportDA() IReportDA {
	_reportDAOnce.Do(func() {
		_reportDA = new(ReportDA)

		learnerReportOverviewCache, err := utils.NewLazyRefreshCache(&utils.LazyRefreshCacheOption{
			RedisKeyPrefix:  RedisKeyPrefixReportLearnerReportOverview,
			Expiration:      constant.ReportQueryCacheExpiration,
			RefreshDuration: constant.ReportQueryCacheRefreshDuration,
			RawQuery:        _reportDA.getLearnerReportOverview})
		if err != nil {
			log.Panic(context.Background(), "create learner report overview cache failed", log.Err(err))
		}

		_reportDA.learnerReportOverviewCache = learnerReportOverviewCache

		learningOutcomeOverviewCache, err := utils.NewLazyRefreshCache(&utils.LazyRefreshCacheOption{
			RedisKeyPrefix:  RedisKeyPrefixReportLearningOutcomeOverview,
			Expiration:      constant.ReportQueryCacheExpiration,
			RefreshDuration: constant.ReportQueryCacheRefreshDuration,
			RawQuery:        _reportDA.getLearningOutcomeOverview})
		if err != nil {
			log.Panic(context.Background(), "create learning outcome overview cache failed", log.Err(err))
		}

		_reportDA.learningOutcomeOverviewCache = learningOutcomeOverviewCache

		teacherUsageOverviewCache, err := utils.NewLazyRefreshCache(&utils.LazyRefreshCacheOption{
			RedisKeyPrefix:  RedisKeyPrefixReportTeacherUsageOverview,
			Expiration:      constant.ReportQueryCacheExpiration,
			RefreshDuration: constant.ReportQueryCacheRefreshDuration,
			RawQuery:        _reportDA.getTeacherUsageOverview})
		if err != nil {
			log.Panic(context.Background(), "create teacher usage overview cache failed", log.Err(err))
		}

		_reportDA.teacherUsageOverviewCache = teacherUsageOverviewCache

		skillCoverageCache, err := utils.NewLazyRefreshCache(&utils.LazyRefreshCacheOption{
			RedisKeyPrefix:  RedisKeyPrefixReportSkillCoverage,
			Expiration:      constant.ReportQueryCacheExpiration,
			RefreshDuration: constant.ReportQueryCacheRefreshDuration,
			RawQuery:        _reportDA.getSkillCoverage})
		if err != nil {
			log.Panic(context.Background(), "create skill coverage cache failed", log.Err(err))
		}

		_reportDA.skillCoverageCache = skillCoverageCache

	})
	return _reportDA
}
