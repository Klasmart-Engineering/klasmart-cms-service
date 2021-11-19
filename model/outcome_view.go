package model

import (
	"context"
	"errors"
	"regexp"

	"strings"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

type ErrValidFailed struct {
	Msg string
}

func (e *ErrValidFailed) Error() string {
	return e.Msg
}

var shortcode3Validate = regexp.MustCompile(`^[A-Z0-9]{3}$`)
var shortcode5Validate = regexp.MustCompile(`^[A-Z0-9]{5}$`)

type OutcomeCreateView struct {
	OutcomeID      string                  `json:"outcome_id"`
	OutcomeName    string                  `json:"outcome_name"`
	Assumed        bool                    `json:"assumed"`
	OrganizationID string                  `json:"organization_id"`
	Program        []string                `json:"program"`
	Subject        []string                `json:"subject"`
	Developmental  []string                `json:"developmental"`
	Skills         []string                `json:"skills"`
	Age            []string                `json:"age"`
	Grade          []string                `json:"grade"`
	Estimated      int                     `json:"estimated_time"`
	Keywords       []string                `json:"keywords"`
	Description    string                  `json:"description"`
	Shortcode      string                  `json:"shortcode,omitempty"`
	Sets           []*OutcomeSetCreateView `json:"sets"`
}

type OutcomeSetCreateView struct {
	SetID   string `json:"set_id" form:"set_id"`
	SetName string `json:"set_name" form:"set_name"`
}

type Milestone struct {
	MilestoneID   string `json:"milestone_id" form:"milestone_id"`
	MilestoneName string `json:"milestone_name" form:"milestone_name"`
}

func (req OutcomeCreateView) ToOutcome(ctx context.Context, op *entity.Operator) (*entity.Outcome, error) {
	outcome := entity.Outcome{
		Name:          req.OutcomeName,
		Assumed:       req.Assumed,
		EstimatedTime: req.Estimated,
		Description:   req.Description,
		Shortcode:     req.Shortcode,
	}

	if len(req.Program) == 0 || len(req.Subject) == 0 {
		log.Warn(ctx, "ToOutcome: program and subject is required", log.Any("op", op), log.Any("req", req))
		return nil, &ErrValidFailed{Msg: "program and subject is required"}
	}

	if !shortcode3Validate.MatchString(req.Shortcode) && !shortcode5Validate.MatchString(req.Shortcode) {
		log.Warn(ctx, "ToOutcome: program and subject is required", log.Any("op", op), log.Any("req", req))
		return nil, &ErrValidFailed{Msg: "shortcode mismatch"}
	}

	_, err := prepareAllNeededName(ctx, op, entity.ExternalOptions{
		OrgIDs:     []string{op.OrgID},
		UsrIDs:     []string{op.UserID},
		ProgIDs:    req.Program,
		SubjectIDs: req.Subject,
		CatIDs:     req.Developmental,
		SubcatIDs:  req.Skills,
		GradeIDs:   req.Grade,
		AgeIDs:     req.Age,
	})
	if err != nil {
		log.Error(ctx, "ToOutcome: prepareAllNeededName failed", log.Err(err), log.Any("op", op), log.Any("req", req))
		return nil, &ErrValidFailed{Msg: "program and subject is required"}
	}

	outcome.Program = strings.Join(req.Program, entity.JoinComma)
	outcome.Subject = strings.Join(req.Subject, entity.JoinComma)
	outcome.Developmental = strings.Join(req.Developmental, entity.JoinComma)
	outcome.Skills = strings.Join(req.Skills, entity.JoinComma)
	outcome.Grade = strings.Join(req.Grade, entity.JoinComma)
	outcome.Age = strings.Join(req.Age, entity.JoinComma)
	outcome.Keywords = strings.Join(req.Keywords, entity.JoinComma)

	outcome.Programs = req.Program
	outcome.Subjects = req.Subject
	outcome.Categories = req.Developmental
	outcome.Subcategories = req.Skills
	outcome.Grades = req.Grade
	outcome.Ages = req.Age

	outcome.Sets = make([]*entity.Set, len(req.Sets))
	for i := range req.Sets {
		set := &entity.Set{
			ID: req.Sets[i].SetID,
		}
		outcome.Sets[i] = set
	}

	return &outcome, nil
}

func (req OutcomeCreateView) ToOutcomeWithID(ctx context.Context, op *entity.Operator, outcomeID string) (*entity.Outcome, error) {
	outcome, err := req.ToOutcome(ctx, op)
	if err != nil {
		return nil, err
	}
	if outcomeID == "" {
		return nil, errors.New("outcomeID invalid")
	}
	outcome.ID = outcomeID
	return outcome, nil
}

type OutcomeCreateResponse struct {
	OutcomeID        string   `json:"outcome_id"`
	OutcomeName      string   `json:"outcome_name"`
	AncestorID       string   `json:"ancestor_id"`
	Shortcode        string   `json:"shortcode"`
	Assumed          bool     `json:"assumed"`
	Program          []string `json:"program"`
	Subject          []string `json:"subject"`
	Developmental    []string `json:"developmental"`
	Skills           []string `json:"skills"`
	Age              []string `json:"age"`
	Grade            []string `json:"grade"`
	EstimatedTime    int      `json:"estimated_time"`
	Keywords         []string `json:"keywords"`
	SourceID         string   `json:"source_id"`
	LockedBy         string   `json:"locked_by"`
	AuthorID         string   `json:"author_id"`
	AuthorName       string   `json:"author_name"`
	OrganizationID   string   `json:"organization_id"`
	OrganizationName string   `json:"organization_name"`
	PublishScope     string   `json:"publish_scope"`
	PublishStatus    string   `json:"publish_status"`
	RejectReason     string   `json:"reject_reason"`
	Description      string   `json:"description"`
	CreatedAt        int64    `json:"created_at"`
	UpdatedAt        int64    `json:"updated_at"`
}

func NewCreateResponse(ctx context.Context, operator *entity.Operator, createView *OutcomeCreateView, outcome *entity.Outcome) OutcomeCreateResponse {
	return OutcomeCreateResponse{
		OutcomeID:        outcome.ID,
		OutcomeName:      createView.OutcomeName,
		AncestorID:       outcome.AncestorID,
		Shortcode:        outcome.Shortcode,
		Assumed:          outcome.Assumed,
		Program:          createView.Program,
		Subject:          createView.Subject,
		Developmental:    createView.Developmental,
		Skills:           createView.Skills,
		Age:              createView.Age,
		Grade:            createView.Grade,
		EstimatedTime:    createView.Estimated,
		Keywords:         createView.Keywords,
		SourceID:         outcome.SourceID,
		LockedBy:         outcome.LockedBy,
		AuthorID:         outcome.AuthorID,
		AuthorName:       outcome.AuthorName,
		OrganizationID:   outcome.OrganizationID,
		OrganizationName: getOrganizationName(ctx, operator, outcome.OrganizationID),
		PublishScope:     outcome.PublishScope,
		PublishStatus:    string(outcome.PublishStatus),
		RejectReason:     outcome.RejectReason,
		Description:      outcome.Description,
		CreatedAt:        outcome.CreateAt,
		UpdatedAt:        outcome.UpdateAt,
	}
}

type OutcomeLockResponse struct {
	OutcomeID string `json:"outcome_id"`
}

type OutcomeIDList struct {
	OutcomeIDs []string `json:"outcome_ids"`
}

type PublishOutcomeReq struct {
	Scope string `json:"scope,omitempty" form:"scope,omitempty"`
}

type OutcomeRejectReq struct {
	RejectReason string `json:"reject_reason"`
}

type OutcomeBulkRejectRequest struct {
	OutcomeIDs   []string `json:"outcome_ids"`
	RejectReason string   `json:"reject_reason"`
}

type Program struct {
	ProgramID   string `json:"program_id"`
	ProgramName string `json:"program_name"`
}

type Subject struct {
	SubjectID   string `json:"subject_id"`
	SubjectName string `json:"subject_name"`
}

type Developmental struct {
	DevelopmentalID   string `json:"developmental_id"`
	DevelopmentalName string `json:"developmental_name"`
}

type Skill struct {
	SkillID   string `json:"skill_id"`
	SkillName string `json:"skill_name"`
}

type Age struct {
	AgeID   string `json:"age_id"`
	AgeName string `json:"age_name"`
}

type Grade struct {
	GradeID   string `json:"grade_id"`
	GradeName string `json:"grade_name"`
}

func getOrganizationName(ctx context.Context, operator *entity.Operator, id string) (name string) {
	ids := []string{id}
	names, err := external.GetOrganizationServiceProvider().GetNameByOrganizationOrSchool(ctx, operator, ids)
	if err != nil {
		log.Error(ctx, "getOrganizationName: GetNameByOrganizationOrSchool failed",
			log.Err(err),
			log.Strings("org_ids", ids))
		return ""
	}
	if len(names) == 0 {
		log.Info(ctx, "getOrganizationName: GetNameByOrganizationOrSchool empty",
			log.Strings("org_ids", ids))
	}
	return names[0]
}

type OutcomeView struct {
	OutcomeID      string                  `json:"outcome_id"`
	AncestorID     string                  `json:"ancestor_id"`
	OutcomeName    string                  `json:"outcome_name"`
	Shortcode      string                  `json:"shortcode"`
	Assumed        bool                    `json:"assumed"`
	LockedBy       string                  `json:"locked_by"`
	LockedLocation []string                `json:"locked_location"`
	LastEditedBy   string                  `json:"last_edited_by"`
	LastEditedAt   int64                   `json:"last_edited_at"`
	AuthorID       string                  `json:"author_id"`
	AuthorName     string                  `json:"author_name"`
	PublishStatus  string                  `json:"publish_status"`
	Program        []Program               `json:"program"`
	Developmental  []Developmental         `json:"developmental"`
	Sets           []*OutcomeSetCreateView `json:"sets"`
	CreatedAt      int64                   `json:"created_at"`
	UpdatedAt      int64                   `json:"update_at"`
}

type SearchOutcomeResponse struct {
	Total int            `json:"total"`
	List  []*OutcomeView `json:"list"`
}

type PublishedOutcomeView struct {
	OutcomeID      string   `json:"outcome_id"`
	OutcomeName    string   `json:"outcome_name"`
	Shortcode      string   `json:"shortcode"`
	ProgramIDs     []string `json:"program_ids"`
	SubjectIDs     []string `json:"subject_ids"`
	CategoryIDs    []string `json:"category_ids"`
	SubcategoryIDs []string `json:"sub_category_ids"`
	GradeIDs       []string `json:"grade_ids"`
	AgeIDs         []string `json:"age_ids"`
}

type SearchPublishedOutcomeResponse struct {
	Total int                     `json:"total"`
	List  []*PublishedOutcomeView `json:"list"`
}

type OutcomeDetailView struct {
	OutcomeID      string                  `json:"outcome_id"`
	OutcomeName    string                  `json:"outcome_name"`
	Shortcode      string                  `json:"shortcode"`
	Description    string                  `json:"description"`
	AuthorID       string                  `json:"author_id"`
	AuthorName     string                  `json:"author_name"`
	Assumed        bool                    `json:"assumed"`
	PublishStatus  string                  `json:"publish_status"`
	RejectReason   string                  `json:"reject_reason"`
	LockedBy       string                  `json:"locked_by"`
	LastEditedBy   string                  `json:"last_edited_by"`
	LastEditedAt   int64                   `json:"last_edited_at"`
	LockedLocation []string                `json:"locked_location"`
	Keywords       []string                `json:"keywords"`
	Program        []Program               `json:"program"`
	Subject        []Subject               `json:"subject"`
	Developmental  []Developmental         `json:"developmental"`
	Skills         []Skill                 `json:"skills"`
	Age            []Age                   `json:"age"`
	Grade          []Grade                 `json:"grade"`
	Sets           []*OutcomeSetCreateView `json:"sets"`
	Milestones     []*Milestone            `json:"milestones"`
	CreatedAt      int64                   `json:"created_at"`
	UpdatedAt      int64                   `json:"update_at"`
}
