package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (cm *ContentModel) GetLessonPlansCanSchedule(ctx context.Context, op *entity.Operator, cond *entity.ContentConditionRequest) (response *entity.GetLessonPlansCanScheduleResponse, err error) {
	response = &entity.GetLessonPlansCanScheduleResponse{
		Data: []*entity.LessonPlanForSchedule{},
	}

	var condOrgContent dbo.Conditions
	if utils.ContainsString(cond.GroupNames, entity.LessonPlanGroupNameOrganizationContent.String()) {
		condOrgContent, err = cm.buildUserContentCondition(ctx, dbo.MustGetDB(ctx), cond, []string{}, op)
		if err != nil {
			return
		}
	}
	needProgramGroup := false
	if utils.ContainsString(cond.GroupNames, entity.LessonPlanGroupNameBadanamuContent.String()) {
		needProgramGroup = true
	}
	if utils.ContainsString(cond.GroupNames, entity.LessonPlanGroupNameMoreFeaturedContent.String()) {
		needProgramGroup = true
	}
	var programGroups []*entity.ProgramGroup
	if needProgramGroup {
		programGroups, err = GetProgramGroupModel().Query(ctx, &da.ProgramGroupQueryCondition{})
		if err != nil {
			return
		}
	}

	response.Total, response.Data, err = da.GetContentDA().GetLessonPlansCanSchedule(ctx, op, cond, condOrgContent, programGroups)
	if err != nil {
		return
	}
	return
}
