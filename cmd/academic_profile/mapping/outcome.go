package mapping

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type OutcomeService struct{}

var ErrOutcomeRepairServiceDirtyData = errors.New("dirt data")

func (o OutcomeService) generateRelations(ctx context.Context, mapper Mapper, outcome *entity.Outcome) (bool, []*entity.Relation, error) {
	programs := strings.Split(outcome.Program, ",")
	subjects := strings.Split(outcome.Subject, ",")
	categories := strings.Split(outcome.Developmental, ",")
	subCategories := strings.Split(outcome.Skills, ",")
	ages := strings.Split(outcome.Age, ",")
	grades := strings.Split(outcome.Grade, ",")
	if len(programs) != 1 {
		log.Error(ctx, "dirty data", log.Any("outcome", outcome))
		return false, nil, ErrOutcomeRepairServiceDirtyData
	}
	if len(subCategories) > 0 && len(categories) != 1 {
		log.Error(ctx, "dirty data", log.Any("outcome", outcome))
		return false, nil, ErrOutcomeRepairServiceDirtyData
	}

	relations := make([]*entity.Relation, 0, len(programs)+len(subjects)+len(categories)+len(subCategories)+len(ages)+len(grades))

	programNeedUpdate := false
	for i := range programs {
		program := mapper.Program(ctx, outcome.OrganizationID, programs[i])
		if program != programs[i] {
			programs[i] = program
			programNeedUpdate = true
		}
		relation := &entity.Relation{
			MasterID:     outcome.ID,
			MasterType:   entity.OutcomeType,
			RelationID:   program,
			RelationType: entity.ProgramType,
		}
		relations = append(relations, relation)
	}
	if programNeedUpdate {
		outcome.Program = strings.Join(programs, ",")
	}

	subjectNeedUpdate := false
	for i := range subjects {
		subject := mapper.Subject(ctx, outcome.OrganizationID, outcome.Program, subjects[i])
		if subject != subjects[i] {
			subjects[i] = subject
			subjectNeedUpdate = true
		}
		relation := &entity.Relation{
			MasterID:     outcome.ID,
			MasterType:   entity.OutcomeType,
			RelationID:   subject,
			RelationType: entity.SubjectType,
		}
		relations = append(relations, relation)
	}
	if subjectNeedUpdate {
		outcome.Subject = strings.Join(subjects, ",")
	}

	categoryNeedUpdate := false
	for i := range categories {
		category := mapper.Category(ctx, outcome.OrganizationID, outcome.Program, categories[i])
		if category != categories[i] {
			categories[i] = category
			categoryNeedUpdate = true
		}
		relation := &entity.Relation{
			MasterID:     outcome.ID,
			MasterType:   entity.OutcomeType,
			RelationID:   category,
			RelationType: entity.CategoryType,
		}
		relations = append(relations, relation)
	}
	if categoryNeedUpdate {
		outcome.Developmental = strings.Join(categories, ",")
	}

	subCategoryNeedUpdate := false
	for i := range subCategories {
		subCategory := mapper.SubCategory(ctx, outcome.OrganizationID, outcome.Program, categories[0], subCategories[i])
		if subCategory != subCategories[i] {
			subCategories[i] = subCategory
			subCategoryNeedUpdate = true
		}
		relation := &entity.Relation{
			MasterID:     outcome.ID,
			MasterType:   entity.OutcomeType,
			RelationID:   subCategory,
			RelationType: entity.SubcategoryType,
		}
		relations = append(relations, relation)
	}
	if subCategoryNeedUpdate {
		outcome.Skills = strings.Join(subCategories, ",")
	}

	ageNeedUpdate := false
	for i := range ages {
		age := mapper.Age(ctx, outcome.OrganizationID, outcome.Program, ages[i])
		if age != ages[i] {
			ages[i] = age
			ageNeedUpdate = true
		}
		relation := &entity.Relation{
			MasterID:     outcome.ID,
			MasterType:   entity.OutcomeType,
			RelationID:   age,
			RelationType: entity.AgeType,
		}
		relations = append(relations, relation)
	}
	if ageNeedUpdate {
		outcome.Age = strings.Join(ages, ",")
	}

	gradeNeedUpdate := false
	for i := range grades {
		grade := mapper.Grade(ctx, outcome.OrganizationID, outcome.Program, grades[i])
		if grade != grades[i] {
			grades[i] = grade
			gradeNeedUpdate = true
		}
		relation := &entity.Relation{
			MasterID:     outcome.ID,
			MasterType:   entity.OutcomeType,
			RelationID:   grade,
			RelationType: entity.GradeType,
		}
		relations = append(relations, relation)
	}
	if gradeNeedUpdate {
		outcome.Grade = strings.Join(grades, ",")
	}

	needUpdate := programNeedUpdate || subjectNeedUpdate || categoryNeedUpdate || subCategoryNeedUpdate || ageNeedUpdate || gradeNeedUpdate
	return needUpdate, relations, nil
}

func (o OutcomeService) Do(ctx context.Context, cliContext *cli.Context, mapper Mapper) error {
	tx := dbo.MustGetDB(ctx)
	querySql := fmt.Sprintf("select id, program, subject, developmental, skills, age, grade from %s where delete_at is null or delete_at = 0", entity.Outcome{}.TableName())
	rows, err := tx.Raw(querySql).Rows()
	if err != nil {
		log.Error(ctx, "select outcome failed", log.Err(err))
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var outcome entity.Outcome
		err = rows.Scan(&outcome.ID, &outcome.Program, &outcome.Subject, &outcome.Developmental, &outcome.Skills, &outcome.Age, &outcome.Grade)
		if err != nil {
			log.Error(ctx, "Scan outcome failed", log.Err(err))
			return err
		}
		needUpdate, relations, err := o.generateRelations(ctx, mapper, &outcome)
		if err == ErrOutcomeRepairServiceDirtyData {
			continue
		}
		if err != nil {
			log.Error(ctx, "mapping failed", log.Err(err), log.Any("outcome", outcome))
			return err
		}
		err = da.GetOutcomeRelationDA().DeleteTx(ctx, tx, []string{outcome.ID})
		if err != nil {
			log.Error(ctx, "delete relation failed", log.Err(err), log.Any("outcome", outcome))
			return err
		}
		err = da.GetOutcomeRelationDA().InsertTx(ctx, tx, relations)
		if err != nil {
			log.Error(ctx, "InsertTx relation failed", log.Err(err), log.Any("relations", relations))
			return err
		}
		if needUpdate {
			updateSql := fmt.Sprintf("update %s set program=?, subject=?, developmental=?, skills=?, age=?, grade=? where id=?", outcome.TableName())
			err = tx.Exec(updateSql, outcome.Program, outcome.Subject, outcome.Developmental, outcome.Skills, outcome.Age, outcome.Grade, outcome.ID).Error
			if err != nil {
				log.Error(ctx, "InsertTx relation failed", log.Err(err), log.Any("outcome", outcome))
				return err
			}
		}
		log.Info(ctx, "outcome is ok")
	}
	return nil
}
