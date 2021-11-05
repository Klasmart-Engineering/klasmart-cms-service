package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (cm *ContentModel) GetLessonPlansCanSchedule(ctx context.Context, op *entity.Operator, condReq *entity.ContentConditionRequest) (lps []*entity.LessonPlanForSchedule, err error) {
	lps = []*entity.LessonPlanForSchedule{}
	searchUserIDs := cm.getRelatedUserID(ctx, condReq.Name, op)
	userContentCondition, err := cm.buildUserContentCondition(ctx, dbo.MustGetDB(ctx), condReq, searchUserIDs, op)
	if err != nil {
		return
	}
	_, contents, err := da.GetContentDA().SearchContentUnSafe(ctx, dbo.MustGetDB(ctx), userContentCondition)
	if err != nil {
		return
	}
	for _, contentInfo := range contents {
		lps = append(lps, &entity.LessonPlanForSchedule{
			ID:        contentInfo.ID,
			Name:      contentInfo.Name,
			GroupName: entity.LessonPlanGroupNameOrganizationContent,
		})
	}

	lps1, err := da.GetContentDA().GetLessonPlansCanSchedule(ctx, op)
	if err != nil {
		return
	}
	lps = append(lps, lps1...)
	return
}
