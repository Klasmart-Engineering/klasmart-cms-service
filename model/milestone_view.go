package model

import (
	"context"
	"sync"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
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
	Name         string                 `json:"milestone_name"`
	Shortcode    string                 `json:"shortcode"`
	Type         entity.TypeOfMilestone `json:"type"`
	Organization *OrganizationView      `json:"organization,omitempty"`
	Author       *AuthorView            `json:"author,omitempty"`
	Outcomes     []*OutcomeView         `json:"outcomes"`
	CreateAt     int64                  `json:"create_at"`
	Program      []*Program             `json:"program"`
	Subject      []*Subject             `json:"subject"`
	Category     []*Category            `json:"category"`
	SubCategory  []*SubCategory         `json:"sub_category"`
	Age          []*Age                 `json:"age"`
	Grade        []*Grade               `json:"grade"`
	Description  string                 `json:"description"`
	Status       string                 `json:"status"`
	LockedBy     string                 `json:"locked_by"`
	AncestorID   string                 `json:"ancestor_id"`
	SourceID     string                 `json:"source_id"`
	LatestID     string                 `json:"latest_id"`
	OutcomeCount int                    `json:"outcome_count"`
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
	if len([]rune(ms.Name)) > constant.MilestoneNameLength {
		return nil, constant.ErrExceededLimit
	}
	if ms.Type == "" {
		ms.Type = entity.CustomMilestoneType
	}
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
	_, err := prepareAllNeededName(ctx, op, entity.ExternalOptions{
		OrgIDs:     []string{op.OrgID},
		UsrIDs:     []string{op.UserID},
		ProgIDs:    ms.ProgramIDs,
		SubjectIDs: ms.SubjectIDs,
		CatIDs:     ms.CategoryIDs,
		SubcatIDs:  ms.SubcategoryIDs,
		GradeIDs:   ms.GradeIDs,
		AgeIDs:     ms.AgeIDs,
	})
	if err != nil {
		log.Error(ctx, "ToMilestone: prepareAllNeededName failed",
			log.Err(err),
			log.Any("op", op),
			log.Any("milestone", ms))
		return nil, err
	}
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
	Milestones []*MilestoneView `json:"milestones"`
	Total      int              `json:"total"`
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

		for _, outcome := range milestones[i].Outcomes {
			orgIDs = append(orgIDs, outcome.OrganizationID)
			authIDs = append(authIDs, outcome.AuthorID)
			prgIDs = append(prgIDs, outcome.Programs...)
			sbjIDs = append(sbjIDs, outcome.Subjects...)
			catIDs = append(catIDs, outcome.Categories...)
			sbcIDs = append(sbcIDs, outcome.Subcategories...)
			grdIDs = append(grdIDs, outcome.Grades...)
			ageIDs = append(ageIDs, outcome.Ages...)
		}
	}
	externalNameMap, err := prepareAllNeededName(ctx, op, entity.ExternalOptions{
		OrgIDs:     orgIDs,
		UsrIDs:     authIDs,
		ProgIDs:    prgIDs,
		SubjectIDs: sbjIDs,
		CatIDs:     catIDs,
		SubcatIDs:  sbcIDs,
		GradeIDs:   grdIDs,
		AgeIDs:     ageIDs,
	})
	if err != nil {
		log.Error(ctx, "fromMilestones: prepareAllNeededName failed",
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
				OrganizationName: externalNameMap.OrgIDMap[milestone.OrganizationID],
			},
			Author: &AuthorView{
				AuthorID:   milestone.AuthorID,
				AuthorName: externalNameMap.UsrIDMap[milestone.AuthorID],
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
		milestoneView.FillAllKindsOfName(externalNameMap.ProgIDMap, externalNameMap.SubjectIDMap,
			externalNameMap.CatIDMap, externalNameMap.SubcatIDMap, externalNameMap.GradeIDMap, externalNameMap.AgeIDMap, milestone)
		milestoneView.Outcomes = make([]*OutcomeView, len(milestone.Outcomes))
		for i, outcome := range milestone.Outcomes {
			milestoneView.Outcomes[i] = buildOutcomeView(ctx, externalNameMap, outcome)
		}
		milestoneViews[i] = &milestoneView
	}
	return milestoneViews, nil
}

func prepareAllNeededName(ctx context.Context, op *entity.Operator, externalOptions entity.ExternalOptions) (
	externalNameMap entity.ExternalNameMap, err error) {

	_organizationIDs := utils.SliceDeduplicationExcludeEmpty(externalOptions.OrgIDs)
	_userIDs := utils.SliceDeduplicationExcludeEmpty(externalOptions.UsrIDs)
	_programIDs := utils.SliceDeduplicationExcludeEmpty(externalOptions.ProgIDs)
	_subjectIDs := utils.SliceDeduplicationExcludeEmpty(externalOptions.SubjectIDs)
	_categoryIDs := utils.SliceDeduplicationExcludeEmpty(externalOptions.CatIDs)
	_subcategoryIDs := utils.SliceDeduplicationExcludeEmpty(externalOptions.SubcatIDs)
	_gradeIDs := utils.SliceDeduplicationExcludeEmpty(externalOptions.GradeIDs)
	_ageIDs := utils.SliceDeduplicationExcludeEmpty(externalOptions.AgeIDs)

	ctxNew, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := new(sync.WaitGroup)

	if len(_organizationIDs) > 0 {
		wg.Add(1)
		go func(ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()
			var ero error
			organizations, ero := external.GetOrganizationServiceProvider().BatchGetNameMap(ctx, op, _organizationIDs)
			if ero != nil {
				log.Error(ctx, "prepareAllNeededName: GetOrganizationServiceProvider failed", log.Err(ero), log.Strings("org", _organizationIDs))
				err = ero
				cancel()
			}
			externalNameMap.OrgIDMap = organizations
		}(ctxNew, cancel)
	} else {
		externalNameMap.OrgIDMap = map[string]string{}
	}

	if len(_userIDs) > 0 {
		wg.Add(1)
		go func(ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()
			var ero error
			users, ero := external.GetUserServiceProvider().BatchGetNameMap(ctx, op, _userIDs)
			if ero != nil {
				log.Error(ctx, "prepareAllNeededName: GetUserServiceProvider failed", log.Err(ero), log.Strings("user", _userIDs))
				err = ero
				cancel()
			}
			externalNameMap.UsrIDMap = users
		}(ctxNew, cancel)
	} else {
		externalNameMap.UsrIDMap = map[string]string{}
	}

	if len(_programIDs) > 0 {
		wg.Add(1)
		go func(ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()
			var ero error
			programs, ero := external.GetProgramServiceProvider().BatchGetNameMap(ctx, op, _programIDs)
			if ero != nil {
				log.Error(ctx, "prepareAllNeededName: GetProgramServiceProvider failed", log.Err(ero), log.Strings("program", _programIDs))
				err = ero
				cancel()
			}
			externalNameMap.ProgIDMap = programs
		}(ctxNew, cancel)
	} else {
		externalNameMap.ProgIDMap = map[string]string{}
	}

	if len(_subjectIDs) > 0 {
		wg.Add(1)
		go func(ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()
			var ero error
			subjects, ero := external.GetSubjectServiceProvider().BatchGetNameMap(ctx, op, _subjectIDs)
			if ero != nil {
				log.Error(ctx, "prepareAllNeededName: GetSubjectServiceProvider failed", log.Err(ero), log.Strings("subject", _subjectIDs))
				err = ero
				cancel()
			}
			externalNameMap.SubjectIDMap = subjects
		}(ctxNew, cancel)
	} else {
		externalNameMap.SubjectIDMap = map[string]string{}
	}

	if len(_categoryIDs) > 0 {
		wg.Add(1)
		go func(ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()
			var ero error
			categories, ero := external.GetCategoryServiceProvider().BatchGetNameMap(ctx, op, _categoryIDs)
			if ero != nil {
				log.Error(ctx, "prepareAllNeededName: GetCategoryServiceProvider failed", log.Err(ero), log.Strings("category", _categoryIDs))
				err = ero
				cancel()
			}
			externalNameMap.CatIDMap = categories
		}(ctxNew, cancel)
	} else {
		externalNameMap.CatIDMap = map[string]string{}
	}

	if len(_subcategoryIDs) > 0 {
		wg.Add(1)
		go func(ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()
			var ero error
			subcategories, ero := external.GetSubCategoryServiceProvider().BatchGetNameMap(ctx, op, _subcategoryIDs)
			if ero != nil {
				log.Error(ctx, "prepareAllNeededName: GetSubCategoryServiceProvider failed", log.Err(ero), log.Strings("subcategory", _subcategoryIDs))
				err = ero
				cancel()
			}
			externalNameMap.SubcatIDMap = subcategories
		}(ctxNew, cancel)
	} else {
		externalNameMap.SubcatIDMap = map[string]string{}
	}

	if len(_gradeIDs) > 0 {
		wg.Add(1)
		go func(ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()
			var ero error
			grades, ero := external.GetGradeServiceProvider().BatchGetNameMap(ctx, op, _gradeIDs)
			if ero != nil {
				log.Error(ctx, "prepareAllNeededName: GetGradeServiceProvider failed", log.Err(ero), log.Strings("grade", _gradeIDs))
				err = ero
				cancel()
			}
			externalNameMap.GradeIDMap = grades
		}(ctxNew, cancel)

	} else {
		externalNameMap.GradeIDMap = map[string]string{}
	}

	if len(_ageIDs) > 0 {
		wg.Add(1)
		go func(ctx context.Context, cancel context.CancelFunc) {
			defer wg.Done()
			var ero error
			ages, ero := external.GetAgeServiceProvider().BatchGetNameMap(ctx, op, _ageIDs)
			if ero != nil {
				log.Error(ctx, "prepareAllNeededName: GetAgeServiceProvider failed", log.Err(ero), log.Strings("age", _ageIDs))
				err = ero
				cancel()
			}
			externalNameMap.AgeIDMap = ages
		}(ctxNew, cancel)
	} else {
		externalNameMap.AgeIDMap = map[string]string{}
	}

	wg.Wait()
	return
}

type MilestoneRejectReq struct {
	RejectReason string `json:"reject_reason"`
}

type MilestoneBulkRejectRequest struct {
	RejectReason string   `json:"reject_reason"`
	MilestoneIDs []string `json:"milestone_ids"`
}
