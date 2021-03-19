package main

import (
	"context"
	"database/sql"
	"fmt"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	iap "gitlab.badanamu.com.cn/calmisland/kidsloop2/cmd/intergrate_academic_profile"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func scheduleMigrate(ctx context.Context, operator *entity.Operator, mapper iap.Mapper, orgID string) ([]string, error) {
	var commands []string
	pager := dbo.Pager{
		Page:     1,
		PageSize: BatchSize,
	}

	for {
		schedules := []*entity.Schedule{}
		total, err := da.GetScheduleDA().Page(ctx, da.ScheduleCondition{
			OrgID: sql.NullString{String: orgID, Valid: true},
			Pager: pager,
		}, &schedules)
		if err != nil {
			return nil, err
		}

		for _, schedule := range schedules {
			command, err := scheduleMapping(ctx, mapper, schedule)
			if err != nil {
				return nil, err
			}

			commands = append(commands, command)
		}

		if len(commands) >= total {
			break
		}

		pager.Page++
	}

	return commands, nil
}

func scheduleMapping(ctx context.Context, mapper iap.Mapper, schedule *entity.Schedule) (string, error) {
	if schedule.ProgramID == "" {
		return "", nil
	}

	programID, err := mapper.Program(ctx, schedule.OrgID, schedule.ProgramID)
	if err != nil {
		return "", err
	}

	subjectID, err := mapper.Subject(ctx, schedule.OrgID, schedule.ProgramID, schedule.SubjectID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("update schedules set program_id='%s', subject_id='%s' where id='%s';",
		programID, subjectID, schedule.ID), nil
}
