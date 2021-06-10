package mapping

import (
	"context"
	"errors"
	"strings"

	"github.com/urfave/cli/v2"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type OutcomeService struct{}

func (o OutcomeService) fetch(ctx context.Context) ([]*entity.Outcome, map[string][]*entity.Relation, error) {
	tx := dbo.MustGetDB(ctx)
	_, outcomes, err := da.GetOutcomeDA().SearchOutcome(ctx, nil, tx, &da.OutcomeCondition{})
	if err != nil {
		return nil, nil, err
	}

	relations, err := da.GetOutcomeRelationDA().SearchTx(ctx, tx, &da.RelationCondition{})
	if err != nil {
		return nil, nil, err
	}
	mapRelations := make(map[string][]*entity.Relation)
	for i := range relations {
		mapRelations[relations[i].MasterID] = append(mapRelations[relations[i].MasterID], relations[i])
	}
	return outcomes, mapRelations, nil
}

func (o OutcomeService) generateRelations(ctx context.Context, mapper Mapper, outcome *entity.Outcome) (bool, []*entity.Relation, error) {
	programs := strings.Split(outcome.Program, ",")
	subjects := strings.Split(outcome.Subject, ",")
	categories := strings.Split(outcome.Developmental, ",")
	subCategories := strings.Split(outcome.Skills, ",")
	ages := strings.Split(outcome.Age, ",")
	grades := strings.Split(outcome.Grade, ",")
	if len(programs) != 1 {
		log.Error(ctx, "dirty data", log.Any("outcome", outcome))
		return false, nil, errors.New("dirty data")
	}
	if len(subCategories) > 0 && len(categories) != 1 {
		log.Error(ctx, "dirty data", log.Any("outcome", outcome))
		return false, nil, errors.New("dirty data")
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

func (o OutcomeService) mappingRelations(ctx context.Context, mapper Mapper, outcome *entity.Outcome, relations []*entity.Relation) ([]*entity.Relation, error) {
	programs := make([]string, 0)
	subjects := make([]string, 0)
	categories := make([]string, 0)
	subCategories := make([]string, 0)
	ages := make([]string, 0)
	grades := make([]string, 0)
	for i := range relations {
		switch relations[i].RelationType {
		case entity.ProgramType:
			programs = append(programs, relations[i].RelationID)
		case entity.SubjectType:
			subjects = append(subjects, relations[i].RelationID)
		case entity.CategoryType:
			categories = append(categories, relations[i].RelationID)
		case entity.SubcategoryType:
			subCategories = append(subCategories, relations[i].RelationID)
		case entity.AgeType:
			ages = append(ages, relations[i].RelationID)
		case entity.GradeType:
			grades = append(grades, relations[i].RelationID)
		}
	}
	if len(programs) != 1 {
		log.Error(ctx, "dirty data", log.Any("outcome", outcome))
		return nil, errors.New("dirty data")
	}
	if len(subCategories) > 0 && len(categories) != 1 {
		log.Error(ctx, "dirty data", log.Any("outcome", outcome))
		return nil, errors.New("dirty data")
	}

	updateRelations := make([]*entity.Relation, 0)
	for i := range relations {
		switch relations[i].RelationType {
		case entity.ProgramType:
			relationID := mapper.Program(ctx, outcome.OrganizationID, relations[i].RelationID)
			if relationID != relations[i].RelationID {
				relations[i].RelationID = relationID
				updateRelations = append(updateRelations, relations[i])
			}
		case entity.SubjectType:
			relationID := mapper.Subject(ctx, outcome.OrganizationID, programs[0], relations[i].RelationID)
			if relationID != relations[i].RelationID {
				relations[i].RelationID = relationID
				updateRelations = append(updateRelations, relations[i])
			}
		case entity.CategoryType:
			relationID := mapper.Category(ctx, outcome.OrganizationID, programs[0], relations[i].RelationID)
			if relationID != relations[i].RelationID {
				relations[i].RelationID = relationID
				updateRelations = append(updateRelations, relations[i])
			}
		case entity.SubcategoryType:
			relationID := mapper.SubCategory(ctx, outcome.OrganizationID, programs[0], categories[0], relations[i].RelationID)
			if relationID != relations[i].RelationID {
				relations[i].RelationID = relationID
				updateRelations = append(updateRelations, relations[i])
			}
		case entity.AgeType:
			relationID := mapper.Age(ctx, outcome.OrganizationID, programs[0], relations[i].RelationID)
			if relationID != relations[i].RelationID {
				relations[i].RelationID = relationID
				updateRelations = append(updateRelations, relations[i])
			}
		case entity.GradeType:
			relationID := mapper.Grade(ctx, outcome.OrganizationID, programs[0], relations[i].RelationID)
			if relationID != relations[i].RelationID {
				relations[i].RelationID = relationID
				updateRelations = append(updateRelations, relations[i])
			}
		}
	}

	return updateRelations, nil
}
func (o OutcomeService) handle(ctx context.Context, mapper Mapper, outcome *entity.Outcome, relations []*entity.Relation) error {
	outcomeNeedUpdate := false
	var insertOrUpdateRelations []*entity.Relation
	var err error
	if relations == nil {
		outcomeNeedUpdate, insertOrUpdateRelations, err = o.generateRelations(ctx, mapper, outcome)
		if err != nil {
			return err
		}
	} else {
		insertOrUpdateRelations, err = o.mappingRelations(ctx, mapper, outcome, relations)
		if err != nil {
			return err
		}
	}
	if outcomeNeedUpdate {
		err = da.GetOutcomeRelationDA().InsertTx(ctx, dbo.MustGetDB(ctx), insertOrUpdateRelations)
		if err != nil {
			log.Error(ctx, "insert relation failed", log.Any("outcome", insertOrUpdateRelations))
			return err
		}
		err = da.GetOutcomeDA().UpdateOutcome(ctx, nil, dbo.MustGetDB(ctx), outcome)
		if err != nil {
			log.Warn(ctx, "update outcome failed", log.Any("outcome", outcome))
		}
		return nil
	}
	err = da.GetOutcomeRelationDA().DeleteTx(ctx, dbo.MustGetDB(ctx), []string{outcome.ID})
	if err != nil {
		log.Error(ctx, "delete relation failed", log.Any("outcome", outcome))
		return err
	}

	err = da.GetOutcomeRelationDA().InsertTx(ctx, dbo.MustGetDB(ctx), insertOrUpdateRelations)
	if err != nil {
		log.Error(ctx, "insert relation failed", log.Any("outcome", insertOrUpdateRelations))
		return err
	}
	return nil
}

func (o OutcomeService) Do(ctx context.Context, cliContext *cli.Context, mapper Mapper) error {
	outcomes, relations, err := o.fetch(ctx)
	if err != nil {
		log.Error(ctx, "fetch failed")
		return err
	}

	for i := range outcomes {
		err := o.handle(ctx, mapper, outcomes[i], relations[outcomes[i].ID])
		if err != nil {
			log.Info(ctx, "handle failed",
				log.Err(err),
				log.Any("outcome", outcomes[i]),
				log.Any("relations", relations[outcomes[i].ID]))
			return err
		}
	}
	return nil
}
