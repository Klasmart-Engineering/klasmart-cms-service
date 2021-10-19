package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (m *reportModel) GetStudentProgressLearnOutcomeAchievement(ctx context.Context, op *entity.Operator, req *entity.LearnOutcomeAchievementRequest) (res *entity.LearnOutcomeAchievementResponse, err error) {
	counts, err := da.GetReportDA().GetStudentProgressLearnOutcomeCountByStudentAndSubject(ctx, req)
	if err != nil {
		return
	}

	_ = counts
	return
}
