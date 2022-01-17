package model

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func (cm *ContentModel) GetLessonPlansCanSchedule(ctx context.Context, op *entity.Operator, cond *entity.ContentConditionRequest) (response *entity.GetLessonPlansCanScheduleResponse, err error) {
	response = &entity.GetLessonPlansCanScheduleResponse{
		Data: []*entity.LessonPlanForSchedule{},
	}

	programs, err := GetProgramModel().GetByOrganization(ctx, op)
	if err != nil {
		return
	}
	var programIDs []string
	for _, program := range programs {
		programIDs = append(programIDs, program.ID)
	}
	if len(cond.ProgramIDs) == 0 {
		cond.ProgramIDs = programIDs
	} else {
		for _, id := range cond.ProgramIDs {
			if !utils.ContainsString(programIDs, id) {
				err = constant.ErrInvalidArgs
				log.Error(ctx,
					"program_id not in current organization",
					log.Any("programIDs", programIDs),
					log.Any("id", id),
				)
				return
			}
		}
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
