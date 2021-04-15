package entity

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type Milestone struct {
	ID             string `gorm:"column:id;primary_key" json:"milestone_id"`
	Name           string `gorm:"column:name" json:"milestone_name"`
	Shortcode      string `gorm:"column:shortcode" json:"shortcode"`
	OrganizationID string `gorm:"column:organization_id" json:"organization_id"`
	AuthorID       string `gorm:"column:author_id" json:"author_id"`
	Description    string `gorm:"column:describe" json:"description"`
	LoCounts       int    `gorm:"column:lo_counts" json:"lo_counts"`

	Status OutcomeStatus `gorm:"column:status" json:"status"`

	LockedBy   string `gorm:"column:locked_by" json:"locked_by"`
	AncestorID string `gorm:"column:ancestor_id" json:"ancestor_id"`
	SourceID   string `gorm:"column:source_id" json:"source_id"`
	LatestID   string `gorm:"column:latest_id" json:"latest_id"`

	CreateAt      int64      `gorm:"column:create_at" json:"created_at"`
	UpdateAt      int64      `gorm:"column:update_at" json:"updated_at"`
	DeleteAt      int64      `gorm:"column:delete_at" json:"deleted_at"`
	Programs      []string   `gorm:"-"`
	Subjects      []string   `gorm:"-"`
	Categories    []string   `gorm:"-"`
	Subcategories []string   `gorm:"-"`
	Grades        []string   `gorm:"-"`
	Ages          []string   `gorm:"-"`
	Outcomes      []*Outcome `gorm:"-" json:"outcomes"`
}

func (Milestone) TableName() string {
	return MilestoneTable
}

func (ms Milestone) CollectAttach() []*Attach {
	attaches := make([]*Attach, len(ms.Programs)+len(ms.Subjects)+len(ms.Categories)+len(ms.Subcategories)+len(ms.Grades)+len(ms.Ages))
	offset := 0
	for i := range ms.Programs {
		attach := Attach{
			MasterID:   ms.ID,
			MasterType: MilestoneType,
			AttachID:   ms.Programs[i],
			AttachType: ProgramType,
		}
		attaches[i] = &attach
	}
	offset += len(ms.Programs)

	for i := range ms.Subjects {
		attach := Attach{
			MasterID:   ms.ID,
			MasterType: MilestoneType,
			AttachID:   ms.Subjects[i],
			AttachType: SubjectType,
		}
		attaches[i+offset] = &attach
	}
	offset += len(ms.Subjects)

	for i := range ms.Categories {
		attach := Attach{
			MasterID:   ms.ID,
			MasterType: MilestoneType,
			AttachID:   ms.Categories[i],
			AttachType: CategoryType,
		}
		attaches[i+offset] = &attach
	}
	offset += len(ms.Categories)

	for i := range ms.Subcategories {
		attach := Attach{
			MasterID:   ms.ID,
			MasterType: MilestoneType,
			AttachID:   ms.Subcategories[i],
			AttachType: SubcategoryType,
		}
		attaches[i+offset] = &attach
	}
	offset += len(ms.Subcategories)

	for i := range ms.Grades {
		attach := Attach{
			MasterID:   ms.ID,
			MasterType: MilestoneType,
			AttachID:   ms.Grades[i],
			AttachType: GradeType,
		}
		attaches[i+offset] = &attach
	}
	offset += len(ms.Grades)

	for i := range ms.Ages {
		attach := Attach{
			MasterID:   ms.ID,
			MasterType: MilestoneType,
			AttachID:   ms.Ages[i],
			AttachType: AgeType,
		}
		attaches[i+offset] = &attach
	}
	offset += len(ms.Ages)
	return attaches
}

func (ms *Milestone) FillAttach(attaches []*Attach) {
	for i := range attaches {
		switch attaches[i].AttachType {
		case ProgramType:
			ms.Programs = append(ms.Programs, attaches[i].AttachID)
		case SubjectType:
			ms.Subjects = append(ms.Subjects, attaches[i].AttachID)
		case CategoryType:
			ms.Categories = append(ms.Categories, attaches[i].AttachID)
		case SubcategoryType:
			ms.Subcategories = append(ms.Subcategories, attaches[i].AttachID)
		case GradeType:
			ms.Grades = append(ms.Grades, attaches[i].AttachID)
		case AgeType:
			ms.Ages = append(ms.Ages, attaches[i].AttachID)
		}
	}
}

func (ms *Milestone) Copy(op *Operator) (*Milestone, error) {
	if ms.Status != OutcomeStatusPublished {
		return nil, constant.ErrOperateNotAllowed
	}
	milestone := &Milestone{
		ID:             utils.NewID(),
		Name:           ms.Name,
		Shortcode:      ms.Shortcode,
		OrganizationID: op.OrgID,
		AuthorID:       op.UserID,
		Description:    ms.Description,
		LoCounts:       ms.LoCounts,

		Status: OutcomeStatusDraft,

		AncestorID: ms.AncestorID,
		SourceID:   ms.ID,
	}
	milestone.SourceID = ms.ID
	milestone.LatestID = milestone.ID
	return milestone, nil
}

func (ms *Milestone) Update(milestone *Milestone) {
	ms.Name = milestone.Name
	ms.Description = milestone.Description
	ms.Programs = milestone.Programs
	ms.Subjects = milestone.Subjects
	ms.Categories = milestone.Categories
	ms.Subcategories = milestone.Subcategories
	ms.Grades = milestone.Grades
	ms.Ages = milestone.Ages
}

func (ms Milestone) OrgAthPrgSbjCatSbcGrdAge(context context.Context, operator *Operator) (
	ctx context.Context, op *Operator, orgIDs, athIDs, prgIDs, sbjIDs, catIDs, sbcIDs, grdIDs, ageIDs []string) {
	ctx = context
	op = operator
	for i := range ms.Outcomes {
		// TODO: unfinished
		fmt.Println(ms.Outcomes[i].Program)
	}
	orgIDs = append(orgIDs, ms.OrganizationID)
	athIDs = append(athIDs, ms.AuthorID)
	prgIDs = append(prgIDs, ms.Programs...)
	sbjIDs = append(sbjIDs, ms.Subjects...)
	catIDs = append(catIDs, ms.Categories...)
	sbcIDs = append(sbcIDs, ms.Subcategories...)
	grdIDs = append(grdIDs, ms.Grades...)
	ageIDs = append(ageIDs, ms.Ages...)
	return
}

type MilestoneOutcome struct {
	ID          int    `gorm:"column:id;primary_key"`
	MilestoneID string `gorm:"column:milestone_id" json:"milestone_id"`
	OutcomeID   string `gorm:"column:outcome_id" json:"outcome_id"`
	CreateAt    int64  `gorm:"column:create_at" json:"created_at"`
	UpdateAt    int64  `gorm:"column:update_at" json:"updated_at"`
	DeleteAt    int64  `gorm:"column:delete_at" json:"deleted_at"`
}

func (MilestoneOutcome) TableName() string {
	return "milestones_outcomes"
}

//type MilestoneAttach struct {
//	MilestoneID string `gorm:"column:milestone_id"`
//	AttachID    string `gorm:"column:attach_id"`
//	AttachType  string `gorm:"column:attach_type"`
//	CreateAt    int64  `gorm:"column:create_at"`
//	UpdateAt    int64  `gorm:"column:update_at"`
//	DeleteAt    int64  `gorm:"column:delete_at"`
//}
//
//func (MilestoneAttach) TableName() string {
//	return AttachMilestoneTable
//}

type MilestoneCondition struct {
	ID             string   `json:"id" form:"id"`
	IDs            []string `json:"ids" form:"ids"`
	Name           string   `json:"name" form:"name"`
	Description    string   `json:"description" form:"description"`
	Shortcode      string   `json:"shortcode" form:"shortcode"`
	AuthorID       string   `json:"-" form:"-"`
	AuthorName     string   `json:"author_name" form:"author_name"`
	Page           string   `json:"page" form:"page"`
	PageSize       string   `json:"page_size" form:"page_size"`
	OrderBy        string   `json:"order_by" form:"order_by"`
	Status         string   `json:"status" form:"status"`
	SearchKey      string   `json:"search_key" form:"search_key"`
	AuthorIDs      []string `json:"-" form:"-"`
	OrganizationID string   `json:"-" form:"-"`
}
