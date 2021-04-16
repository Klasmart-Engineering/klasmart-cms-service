package api

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
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
	MilestoneID  string            `json:"milestone_id,omitempty"`
	Name         string            `json:"milestone_name,omitempty"`
	Shortcode    string            `json:"shortcode,omitempty"`
	Organization *OrganizationView `json:"organization,omitempty"`
	Author       *AuthorView       `json:"organization,omitempty"`
	Outcomes     []*OutcomeView    `json:"outcomes,omitempty"`
	CreateAt     int64             `json:"create_at,omitempty"`
	Program      []*Program        `json:"program,omitempty"`
	Subject      []*Subject        `json:"subject,omitempty"`
	Category     []*Category       `json:"category,omitempty"`
	SubCategory  []*SubCategory    `json:"sub_category,omitempty"`
	Age          []*Age            `json:"age,omitempty"`
	Grade        []*Grade          `json:"grade,omitempty"`
	Description  string            `json:"description,omitempty"`
	Status       string            `json:"status,omitempty"`
	LockedBy     string            `json:"locked_by,omitempty"`
	AncestorID   string            `json:"ancestor_id,omitempty"`
	SourceID     string            `json:"source_id,omitempty"`
	LatestID     string            `json:"latest_id,omitempty"`
	OutcomeCount int               `json:"outcome_count,omitempty"`

	ProgramIDs         []string `json:"program_ids,omitempty"`
	SubjectIDs         []string `json:"subject_ids,omitempty"`
	CategoryIDs        []string `json:"category_ids,omitempty"`
	SubcategoryIDs     []string `json:"subcategory_ids,omitempty"`
	GradeIDs           []string `json:"grade_ids,omitempty"`
	AgeIDs             []string `json:"age_ids,omitempty"`
	OutcomeAncestorIDs []string `json:"outcome_ancestor_ids,omitempty"`
}

func (ms *MilestoneView) toMilestone(op *entity.Operator) *entity.Milestone {
	milestone := &entity.Milestone{
		ID:             ms.MilestoneID,
		Name:           ms.Name,
		Shortcode:      ms.Shortcode,
		OrganizationID: op.OrgID,
		AuthorID:       op.UserID,
		Description:    ms.Description,

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
	return milestone
}

func (ms *MilestoneView) fillAllKindsOfName(program, subject, category, subCategory, grade, age map[string]string, milestone *entity.Milestone) {
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

func (ms *MilestoneView) fillOutcomeView(org, author, program, subject, category, subCategory, grade, age map[string]string, outcome []*entity.Outcome) {
	ms.Outcomes = make([]*OutcomeView, len(outcome))
	for i := range outcome {
		ms.Outcomes[i] = buildOutcomeView(org, author, program, subject, category, subCategory, grade, age, outcome[i])
	}
}

type MilestoneList struct {
	IDs []string `json:"ids"`
}
type MilestoneSearchResponse struct {
	Total      int              `json:"total"`
	Milestones []*MilestoneView `json:"milestones"`
}

func fromMilestones(ctx context.Context, op *entity.Operator, milestones []*entity.Milestone) ([]*MilestoneView, error) {
	var orgIDs, authIDs, prgIDs, sbjIDs, catIDs, sbcIDs, grdIDs, ageIDs []string
	for i := range milestones {
		ms := milestones[i]
		orgIDs = append(orgIDs, ms.OrganizationID)
		authIDs = append(authIDs, ms.AuthorID)
		prgIDs = append(prgIDs, ms.Programs...)
		sbjIDs = append(sbjIDs, ms.Subjects...)
		catIDs = append(catIDs, ms.Categories...)
		sbcIDs = append(sbcIDs, ms.Subcategories...)
		grdIDs = append(grdIDs, ms.Grades...)
		ageIDs = append(ageIDs, ms.Ages...)
	}
	orgs, authors, prds, sbjs, cats, sbcs, grds, ages, err := external.GetOrganizationServiceProvider().
		OrgAthPrgSjtCtgSubCtgGrdAge(ctx, op, orgIDs, authIDs, prgIDs, sbjIDs, catIDs, sbcIDs, grdIDs, ageIDs)
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
	for i := range milestones {
		milestone := milestones[i]
		milestoneView := MilestoneView{
			MilestoneID: milestone.ID,
			Name:        milestone.Name,
			Shortcode:   milestone.Shortcode,
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
		milestoneView.fillAllKindsOfName(prds, sbjs, cats, sbcs, grds, ages, milestone)
		milestoneView.fillOutcomeView(orgs, authors, prds, sbjs, cats, sbcs, grds, ages, milestone.Outcomes)
		milestoneViews[i] = &milestoneView
	}
	return milestoneViews, nil
}
