package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
)

type Category struct {
	CategoryID   string `json:"category_id,omitempty"`
	CategoryName string `json:"category_name,omitempty"`
}

type SubCategory struct {
	SubCategoryID   string `json:"sub_category_id,omitempty"`
	SubCategoryName string `json:"sub_category_name,omitempty"`
}

type OrganizationView struct {
	OrganizationID   string `json:"organization_id,omitempty"`
	OrganizationName string `json:"organization_name,omitempty"`
}

type AuthorView struct {
	AuthorID   string `json:"author_id,omitempty"`
	AuthorName string `json:"author_name,omitempty"`
}

type MilestoneView struct {
	MilestoneID  string                 `json:"milestone_id,omitempty"`
	Name         string                 `json:"milestone_name,omitempty"`
	Shortcode    string                 `json:"shortcode,omitempty"`
	Type         entity.TypeOfMilestone `json:"type"`
	Organization *OrganizationView      `json:"organization,omitempty"`
	Author       *AuthorView            `json:"author,omitempty"`
	Outcomes     []*OutcomeView         `json:"outcomes"`
	CreateAt     int64                  `json:"create_at,omitempty"`
	Program      []*Program             `json:"program"`
	Subject      []*Subject             `json:"subject"`
	Category     []*Category            `json:"category"`
	SubCategory  []*SubCategory         `json:"sub_category"`
	Age          []*Age                 `json:"age"`
	Grade        []*Grade               `json:"grade"`
	Description  string                 `json:"description,omitempty"`
	Status       string                 `json:"status,omitempty"`
	LockedBy     string                 `json:"locked_by,omitempty"`
	AncestorID   string                 `json:"ancestor_id,omitempty"`
	SourceID     string                 `json:"source_id,omitempty"`
	LatestID     string                 `json:"latest_id,omitempty"`
	OutcomeCount int                    `json:"outcome_count,omitempty"`
	WithPublish  bool                   `json:"with_publish,omitempty"`

	ProgramIDs         []string `json:"program_ids,omitempty"`
	SubjectIDs         []string `json:"subject_ids,omitempty"`
	CategoryIDs        []string `json:"category_ids,omitempty"`
	SubcategoryIDs     []string `json:"subcategory_ids,omitempty"`
	GradeIDs           []string `json:"grade_ids,omitempty"`
	AgeIDs             []string `json:"age_ids,omitempty"`
	OutcomeAncestorIDs []string `json:"outcome_ancestor_ids,omitempty"`
}

func (ms *MilestoneView) ToMilestone(ctx context.Context, op *entity.Operator) (*entity.Milestone, error) {
	milestone := &entity.Milestone{
		ID:             ms.MilestoneID,
		Name:           ms.Name,
		Shortcode:      ms.Shortcode,
		OrganizationID: op.OrgID,
		AuthorID:       op.UserID,
		Description:    ms.Description,
		Type:           ms.Type,

		Status: entity.OutcomeStatus(ms.Status),

		LockedBy:   ms.LockedBy,
		AncestorID: ms.AncestorID,
		SourceID:   ms.SourceID,
		LatestID:   ms.LatestID,

		Programs:      ms.ProgramIDs,
		Subjects:      ms.SubjectIDs,
		Categories:    ms.CategoryIDs,
		Subcategories: ms.SubcategoryIDs,
		Grades:        ms.GradeIDs,
		Ages:          ms.AgeIDs,
	}
	if len(ms.ProgramIDs) == 0 || len(ms.SubjectIDs) == 0 {
		log.Warn(ctx, "ToMilestone: program and subject is required", log.Any("op", op), log.Any("milestone", ms))
		return nil, &ErrValidFailed{Msg: "program and subject is required"}
	}
	// TODO: just for test
	//_, _, _, _, _, _, _, _, err := prepareAllNeededName(ctx, op, []string{op.OrgID}, []string{op.UserID},
	//	ms.ProgramIDs, ms.SubjectIDs, ms.CategoryIDs, ms.SubcategoryIDs, ms.GradeIDs, ms.AgeIDs)
	//if err != nil {
	//	log.Error(ctx, "ToMilestone: prepareAllNeededName failed",
	//		log.Err(err),
	//		log.Any("op", op),
	//		log.Any("milestone", ms))
	//	return nil, err
	//}
	return milestone, nil
}

func (ms *MilestoneView) FillAllKindsOfName(program, subject, category, subCategory, grade, age map[string]string, milestone *entity.Milestone) {
	ms.Program = make([]*Program, len(milestone.Programs))
	for i := range milestone.Programs {
		pView := Program{
			ProgramID:   milestone.Programs[i],
			ProgramName: program[milestone.Programs[i]],
		}
		ms.Program[i] = &pView
	}

	ms.Subject = make([]*Subject, len(milestone.Subjects))
	for i := range milestone.Subjects {
		sView := Subject{
			SubjectID:   milestone.Subjects[i],
			SubjectName: subject[milestone.Subjects[i]],
		}
		ms.Subject[i] = &sView
	}

	ms.Category = make([]*Category, len(milestone.Categories))
	for i := range milestone.Categories {
		cView := Category{
			CategoryID:   milestone.Categories[i],
			CategoryName: category[milestone.Categories[i]],
		}
		ms.Category[i] = &cView
	}

	ms.SubCategory = make([]*SubCategory, len(milestone.Subcategories))
	for i := range milestone.Subcategories {
		scView := SubCategory{
			SubCategoryID:   milestone.Subcategories[i],
			SubCategoryName: subCategory[milestone.Subcategories[i]],
		}
		ms.SubCategory[i] = &scView
	}

	ms.Grade = make([]*Grade, len(milestone.Grades))
	for i := range milestone.Grades {
		gView := Grade{
			GradeID:   milestone.Grades[i],
			GradeName: grade[milestone.Grades[i]],
		}
		ms.Grade[i] = &gView
	}

	ms.Age = make([]*Age, len(milestone.Ages))
	for i := range milestone.Ages {
		aView := Age{
			AgeID:   milestone.Ages[i],
			AgeName: age[milestone.Ages[i]],
		}
		ms.Age[i] = &aView
	}
}

type MilestoneList struct {
	IDs []string `json:"ids"`
}
type MilestoneSearchResponse struct {
	Total      int              `json:"total"`
	Milestones []*MilestoneView `json:"milestones"`
}

func FromMilestones(ctx context.Context, op *entity.Operator, milestones []*entity.Milestone) ([]*MilestoneView, error) {
	var orgIDs, authIDs, prgIDs, sbjIDs, catIDs, sbcIDs, grdIDs, ageIDs []string
	for i := range milestones {
		orgIDs = append(orgIDs, milestones[i].OrganizationID)
		authIDs = append(authIDs, milestones[i].AuthorID)
		prgIDs = append(prgIDs, milestones[i].Programs...)
		sbjIDs = append(sbjIDs, milestones[i].Subjects...)
		catIDs = append(catIDs, milestones[i].Categories...)
		sbcIDs = append(sbcIDs, milestones[i].Subcategories...)
		grdIDs = append(grdIDs, milestones[i].Grades...)
		ageIDs = append(ageIDs, milestones[i].Ages...)
	}
	orgs, authors, prds, sbjs, cats, sbcs, grds, ages, err := prepareAllNeededName(ctx, op, orgIDs, authIDs, prgIDs, sbjIDs, catIDs, sbcIDs, grdIDs, ageIDs)
	if err != nil {
		log.Error(ctx, "fromMilestones: OrgAthPrgSjtCtgSubCtgGrdAge failed",
			log.Err(err),
			log.Any("op", op),
			log.Strings("org", orgIDs),
			log.Strings("author", authIDs),
			log.Strings("program", prgIDs),
			log.Strings("subject", sbjIDs),
			log.Strings("category", catIDs),
			log.Strings("subcategory", sbcIDs),
			log.Strings("grade", grdIDs),
			log.Strings("age", ageIDs))
		return nil, err
	}
	milestoneViews := make([]*MilestoneView, len(milestones))
	for i, milestone := range milestones {
		milestoneView := MilestoneView{
			MilestoneID: milestone.ID,
			Name:        milestone.Name,
			Shortcode:   milestone.Shortcode,
			Type:        milestone.Type,
			Organization: &OrganizationView{
				OrganizationID:   milestone.OrganizationID,
				OrganizationName: orgs[milestone.OrganizationID],
			},
			Author: &AuthorView{
				AuthorID:   milestone.AuthorID,
				AuthorName: authors[milestone.AuthorID],
			},
			OutcomeCount: milestone.LoCounts,
			CreateAt:     milestone.CreateAt,
			Description:  milestone.Description,
			Status:       string(milestone.Status),
			LockedBy:     milestone.LockedBy,
			AncestorID:   milestone.AncestorID,
			SourceID:     milestone.SourceID,
			LatestID:     milestone.LatestID,
		}
		milestoneView.FillAllKindsOfName(prds, sbjs, cats, sbcs, grds, ages, milestone)
		milestoneView.Outcomes = make([]*OutcomeView, len(milestone.Outcomes))
		for i, outcome := range milestone.Outcomes {
			milestoneView.Outcomes[i] = buildOutcomeView(orgs, authors, prds, sbjs, cats, sbcs, grds, ages, outcome)
		}
		milestoneViews[i] = &milestoneView
	}
	return milestoneViews, nil
}

func prepareAllNeededName(ctx context.Context, op *entity.Operator,
	organizationIDs, authorIDs, programIDs, subjectIDs, categoryIDs, subCategoryIDs, gradeIDs, ageIDs []string) (
	organizations, authors, programs, subjects, categories, subcategories, grades, ages map[string]string, err error) {

	_organizationIDs := utils.SliceDeduplicationExcludeEmpty(organizationIDs)
	_authorIDs := utils.SliceDeduplicationExcludeEmpty(authorIDs)
	_programIDs := utils.SliceDeduplicationExcludeEmpty(programIDs)
	_subjectIDs := utils.SliceDeduplicationExcludeEmpty(subjectIDs)
	_categoryIDs := utils.SliceDeduplicationExcludeEmpty(categoryIDs)
	_subcategoryIDs := utils.SliceDeduplicationExcludeEmpty(subCategoryIDs)
	_gradeIDs := utils.SliceDeduplicationExcludeEmpty(gradeIDs)
	_ageIDs := utils.SliceDeduplicationExcludeEmpty(ageIDs)

	ctxNew, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := new(sync.WaitGroup)

	if len(_organizationIDs) > 0 {
		wg.Add(1)
		go func(ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()
			var ero error
			organizations, ero = external.GetOrganizationServiceProvider().BatchGetNameMap(ctx, op, _organizationIDs)
			if ero != nil {
				log.Error(ctx, "prepareAllNeededName: GetOrganizationServiceProvider failed", log.Err(ero), log.Strings("org", _organizationIDs))
				err = ero
				cancel()
			}
		}(ctxNew, cancel)
	} else {
		organizations = map[string]string{}
	}

	if len(_authorIDs) > 0 {
		wg.Add(1)
		go func(ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()
			var ero error
			authors, ero = external.GetUserServiceProvider().BatchGetNameMap(ctx, op, _authorIDs)
			if ero != nil {
				log.Error(ctx, "prepareAllNeededName: GetUserServiceProvider failed", log.Err(ero), log.Strings("author", _authorIDs))
				err = ero
				cancel()
			}
		}(ctxNew, cancel)
	} else {
		authors = map[string]string{}
	}

	if len(_programIDs) > 0 {
		wg.Add(1)
		go func(ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()
			var ero error
			programs, ero = external.GetProgramServiceProvider().BatchGetNameMap(ctx, op, _programIDs)
			if ero != nil {
				log.Error(ctx, "prepareAllNeededName: GetProgramServiceProvider failed", log.Err(ero), log.Strings("program", _programIDs))
				err = ero
				cancel()
			}
		}(ctxNew, cancel)
	} else {
		programs = map[string]string{}
	}

	if len(_subjectIDs) > 0 {
		wg.Add(1)
		go func(ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()
			var ero error
			subjects, ero = external.GetSubjectServiceProvider().BatchGetNameMap(ctx, op, _subjectIDs)
			if ero != nil {
				log.Error(ctx, "prepareAllNeededName: GetSubjectServiceProvider failed", log.Err(ero), log.Strings("subject", _subjectIDs))
				err = ero
				cancel()
			}
		}(ctxNew, cancel)
	} else {
		subjects = map[string]string{}
	}

	if len(_categoryIDs) > 0 {
		wg.Add(1)
		go func(ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()
			var ero error
			categories, ero = external.GetCategoryServiceProvider().BatchGetNameMap(ctx, op, _categoryIDs)
			if ero != nil {
				log.Error(ctx, "prepareAllNeededName: GetCategoryServiceProvider failed", log.Err(ero), log.Strings("category", _categoryIDs))
				err = ero
				cancel()
			}
		}(ctxNew, cancel)
	} else {
		categories = map[string]string{}
	}

	if len(_subcategoryIDs) > 0 {
		wg.Add(1)
		go func(ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()
			var ero error
			subcategories, ero = external.GetSubCategoryServiceProvider().BatchGetNameMap(ctx, op, _subcategoryIDs)
			if ero != nil {
				log.Error(ctx, "prepareAllNeededName: GetSubCategoryServiceProvider failed", log.Err(ero), log.Strings("subcategory", _subcategoryIDs))
				err = ero
				cancel()
			}
		}(ctxNew, cancel)
	} else {
		subcategories = map[string]string{}
	}

	if len(_gradeIDs) > 0 {
		wg.Add(1)
		go func(ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()
			var ero error
			grades, ero = external.GetGradeServiceProvider().BatchGetNameMap(ctx, op, _gradeIDs)
			if ero != nil {
				log.Error(ctx, "prepareAllNeededName: GetGradeServiceProvider failed", log.Err(ero), log.Strings("grade", _gradeIDs))
				err = ero
				cancel()
			}
		}(ctxNew, cancel)

	} else {
		grades = map[string]string{}
	}

	if len(_ageIDs) > 0 {
		wg.Add(1)
		go func(ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()
			var ero error
			ages, ero = external.GetAgeServiceProvider().BatchGetNameMap(ctx, op, _ageIDs)
			if ero != nil {
				log.Error(ctx, "prepareAllNeededName: GetAgeServiceProvider failed", log.Err(ero), log.Strings("age", _ageIDs))
				err = ero
				cancel()
			}
		}(ctxNew, cancel)
	} else {
		ages = map[string]string{}
	}

	wg.Wait()
	return
}
