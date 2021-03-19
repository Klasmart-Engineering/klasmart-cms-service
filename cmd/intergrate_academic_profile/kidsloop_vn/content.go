package main

import (
	"context"
	"fmt"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	iap "gitlab.badanamu.com.cn/calmisland/kidsloop2/cmd/intergrate_academic_profile"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

func contentMigrate(ctx context.Context, operator *entity.Operator, mapper iap.Mapper, orgID string) ([]string, error) {
	var commands []string
	pager := utils.Pager{
		PageIndex: 1,
		PageSize:  BatchSize,
	}

	for {
		total, contents, err := da.GetContentDA().SearchContent(ctx, dbo.MustGetDB(ctx), da.ContentCondition{
			Org:   orgID,
			Pager: pager,
		})
		if err != nil {
			return nil, err
		}

		for _, content := range contents {
			command, err := contentMapping(ctx, mapper, content)
			if err != nil {
				return nil, err
			}

			commands = append(commands, command)
		}

		if len(commands) >= total {
			break
		}

		pager.PageIndex++
	}

	return commands, nil
}

func contentMapping(ctx context.Context, mapper iap.Mapper, content *entity.Content) (string, error) {
	programID, err := mapper.Program(ctx, content.Org, content.Program)
	if err != nil {
		return "", err
	}

	var subjectID string
	subjectIDs := strings.Split(content.Subject, ",")
	if len(subjectIDs) > 0 {
		subjectID, err = mapper.Subject(ctx, content.Org, content.Program, subjectIDs[0])
		if err != nil {
			return "", err
		}
	}

	categoryID, err := mapper.Category(ctx, content.Org, content.Program, content.Developmental)
	if err != nil {
		return "", err
	}

	var subCategoryID string
	subCategoryIDs := strings.Split(content.Skills, ",")
	if len(subCategoryIDs) > 0 {
		ids := make([]string, len(subCategoryIDs))
		for index, id := range subCategoryIDs {
			ids[index], err = mapper.SubCategory(ctx, content.Org, content.Program, content.Developmental, id)
			if err != nil {
				return "", err
			}
		}

		subCategoryID = strings.Join(utils.SliceDeduplication(ids), ",")
	}

	var ageID string
	ageIDs := strings.Split(content.Age, ",")
	if len(ageIDs) > 0 {
		ids := make([]string, len(ageIDs))
		for index, id := range ageIDs {
			ids[index], err = mapper.Age(ctx, content.Org, content.Program, id)
			if err != nil {
				return "", err
			}
		}

		ageID = strings.Join(utils.SliceDeduplication(ids), ",")
	}

	var gradeID string
	gradeIDs := strings.Split(content.Grade, ",")
	if len(gradeIDs) > 0 {
		ids := make([]string, len(gradeIDs))
		for index, id := range gradeIDs {
			ids[index], err = mapper.Grade(ctx, content.Org, content.Program, id)
			if err != nil {
				return "", err
			}
		}

		gradeID = strings.Join(utils.SliceDeduplication(ids), ",")
	}

	return fmt.Sprintf("update cms_contents set program='%s', subject='%s', developmental='%s', skills='%s', age='%s', grade='%s' where id='%s';",
		programID, subjectID, categoryID, subCategoryID, ageID, gradeID, content.ID), nil
}
