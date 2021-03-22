package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	iap "gitlab.badanamu.com.cn/calmisland/kidsloop2/cmd/intergrate_academic_profile"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

func outcomeMigrate(ctx context.Context, operator *entity.Operator, mapper iap.Mapper, orgID string) ([]string, error) {
	var commands []string
	pager := dbo.Pager{
		Page:     1,
		PageSize: BatchSize,
	}

	for {
		total, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, dbo.MustGetDB(ctx), &da.OutcomeCondition{
			OrganizationID: sql.NullString{String: orgID, Valid: true},
			Pager:          pager,
		})
		if err != nil {
			return nil, err
		}

		for _, outcome := range outcomes {
			command, err := outcomeMapping(ctx, mapper, outcome)
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

func outcomeMapping(ctx context.Context, mapper iap.Mapper, outcome *entity.Outcome) (string, error) {
	if outcome.Program == "" {
		return "", nil
	}

	programID, err := mapper.Program(ctx, outcome.OrganizationID, outcome.Program)
	if err != nil {
		return "", err
	}

	var subjectID string
	subjectIDs := strings.Split(outcome.Subject, ",")
	if len(subjectIDs) > 0 {
		subjectID, err = mapper.Subject(ctx, outcome.OrganizationID, outcome.Program, subjectIDs[0])
		if err != nil {
			return "", err
		}
	}

	categoryID, err := mapper.Category(ctx, outcome.OrganizationID, outcome.Program, outcome.Developmental)
	if err != nil {
		return "", err
	}

	var subCategoryID string
	subCategoryIDs := strings.Split(outcome.Skills, ",")
	if len(subCategoryIDs) > 0 {
		ids := make([]string, len(subCategoryIDs))
		for index, id := range subCategoryIDs {
			ids[index], err = mapper.SubCategory(ctx, outcome.OrganizationID, outcome.Program, outcome.Developmental, id)
			if err != nil {
				return "", err
			}
		}

		subCategoryID = strings.Join(utils.SliceDeduplication(ids), ",")
	}

	var ageID string
	ageIDs := strings.Split(outcome.Age, ",")
	if len(ageIDs) > 0 {
		ids := make([]string, len(ageIDs))
		for index, id := range ageIDs {
			ids[index], err = mapper.Age(ctx, outcome.OrganizationID, outcome.Program, id)
			if err != nil {
				return "", err
			}
		}

		ageID = strings.Join(utils.SliceDeduplication(ids), ",")
	}

	var gradeID string
	gradeIDs := strings.Split(outcome.Grade, ",")
	if len(gradeIDs) > 0 {
		ids := make([]string, len(gradeIDs))
		for index, id := range gradeIDs {
			ids[index], err = mapper.Grade(ctx, outcome.OrganizationID, outcome.Program, id)
			if err != nil {
				return "", err
			}
		}

		gradeID = strings.Join(utils.SliceDeduplication(ids), ",")
	}

	return fmt.Sprintf("update learning_outcomes set program='%s', `subject`='%s', developmental='%s', skills='%s', age='%s', grade='%s' where id='%s';",
		programID, subjectID, categoryID, subCategoryID, ageID, gradeID, outcome.ID), nil
}
