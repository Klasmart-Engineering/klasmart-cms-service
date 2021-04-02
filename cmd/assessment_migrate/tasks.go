package main

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func handleTeacherIDs(ctx context.Context, tx *dbo.DBContext) error {
	var (
		assessments []*entity.Assessment
		cond = &da.QueryAssessmentConditions{
			OrgID:                   nil,
			Status:                  nil,
			ScheduleIDs:             nil,
			TeacherIDs:              nil,
			AllowTeacherIDs:         nil,
			TeacherIDAndStatusPairs: nil,
			OrderBy:                 nil,
			Page:                    0,
			PageSize:                0,
		}
	)
	if err := da.GetAssessmentDA().QueryTx(ctx, tx, cond, &assessments); err != nil {
		return err
	}
	return nil
}
