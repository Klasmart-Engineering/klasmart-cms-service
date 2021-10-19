package da

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type IStudentProgressLearnOutcomeAchievement interface {
	GetStudentProgressLearnOutcomeCountByStudentAndSubject(ctx context.Context, req *entity.LearnOutcomeAchievementRequest) (res []*entity.StudentProgressLearnOutcomeCount, err error)
}

func (r *ReportDA) GetStudentProgressLearnOutcomeCountByStudentAndSubject(ctx context.Context, req *entity.LearnOutcomeAchievementRequest) (res []*entity.StudentProgressLearnOutcomeCount, err error) {

	return
}
