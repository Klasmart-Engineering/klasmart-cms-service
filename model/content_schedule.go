package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (cm *ContentModel) GetLessonPlansCanSchedule(ctx context.Context, op *entity.Operator) (lps []*entity.LessonPlanForSchedule, err error) {
	lps = []*entity.LessonPlanForSchedule{}
	userContentCondition, err := cm.buildUserContentCondition(ctx, dbo.MustGetDB(ctx), &entity.ContentConditionRequest{
		OrderBy: "create_at",
	}, []string{}, op)
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
	mPG, err := GetProgramGroupModel().QueryMap(ctx, &da.ProgramGroupQueryCondition{})
	if err != nil {
		return
	}
	for _, lp := range lps1 {
		_, ok := mPG[lp.ProgramID]
		if ok {
			lp.GroupName = entity.LessonPlanGroupNameBadanamuContent
		} else {
			lp.GroupName = entity.LessonPlanGroupNameMoreFeaturedContent
		}
	}

	lps = append(lps, lps1...)
	return
}
