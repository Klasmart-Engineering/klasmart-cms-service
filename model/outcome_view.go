package model

import (
	"context"
	"errors"
	//"gitlab.badanamu.com.cn/calmisland/kidsloop2/api"

	//"gitlab.badanamu.com.cn/calmisland/kidsloop2/api"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

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

func (req OutcomeCreateView) ToOutcome() (*entity.Outcome, error) {
	outcome := entity.Outcome{
		Name:          req.OutcomeName,
		Assumed:       req.Assumed,
		EstimatedTime: req.Estimated,
		Description:   req.Description,
		Shortcode:     req.Shortcode,
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

func (req OutcomeCreateView) ToOutcomeWithID(outcomeID string) (*entity.Outcome, error) {
	outcome, err := req.ToOutcome()
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

type OutcomeView struct {
	OutcomeID        string                  `json:"outcome_id"`
	OutcomeName      string                  `json:"outcome_name"`
	AncestorID       string                  `json:"ancestor_id"`
	Shortcode        string                  `json:"shortcode"`
	Assumed          bool                    `json:"assumed"`
	Program          []Program               `json:"program"`
	Subject          []Subject               `json:"subject"`
	Developmental    []Developmental         `json:"developmental"`
	Skills           []Skill                 `json:"skills"`
	Age              []Age                   `json:"age"`
	Grade            []Grade                 `json:"grade"`
	EstimatedTime    int                     `json:"estimated_time"`
	Keywords         []string                `json:"keywords"`
	SourceID         string                  `json:"source_id"`
	LatestID         string                  `json:"latest_id"`
	LockedBy         string                  `json:"locked_by"`
	AuthorID         string                  `json:"author_id"`
	AuthorName       string                  `json:"author_name"`
	OrganizationID   string                  `json:"organization_id"`
	OrganizationName string                  `json:"organization_name"`
	PublishScope     string                  `json:"publish_scope"`
	PublishStatus    string                  `json:"publish_status"`
	RejectReason     string                  `json:"reject_reason"`
	Description      string                  `json:"description"`
	CreatedAt        int64                   `json:"created_at"`
	UpdatedAt        int64                   `json:"update_at"`
	Sets             []*OutcomeSetCreateView `json:"sets"`
}

type OutcomeSearchResponse struct {
	Total int            `json:"total"`
	List  []*OutcomeView `json:"list"`
}

func NewSearchResponse(ctx context.Context, op *entity.Operator, total int, outcomes []*entity.Outcome) (*OutcomeSearchResponse, error) {
	views, err := FillOutcomeViews(ctx, op, outcomes)
	if err != nil {
		log.Error(ctx, "NewSearchResponse: NewSearchResponse failed",
			log.Any("op", op),
			log.Any("outcome", outcomes))
		return nil, err
	}
	return &OutcomeSearchResponse{
		Total: total,
		List:  views,
	}, nil
}

func FillOutcomeViews(ctx context.Context, operator *entity.Operator, outcomes []*entity.Outcome) ([]*OutcomeView, error) {
	outcomeViews := make([]*OutcomeView, len(outcomes))
	var orgIDs, authIDs, prgIDs, sbjIDs, catIDs, sbcIDs, grdIDs, ageIDs []string
	for i := range outcomes {
		orgIDs = append(orgIDs, outcomes[i].OrganizationID)
		authIDs = append(authIDs, outcomes[i].AuthorID)
		prgIDs = append(prgIDs, outcomes[i].Programs...)
		sbjIDs = append(sbjIDs, outcomes[i].Subjects...)
		catIDs = append(catIDs, outcomes[i].Categories...)
		sbcIDs = append(sbcIDs, outcomes[i].Subcategories...)
		grdIDs = append(grdIDs, outcomes[i].Grades...)
		ageIDs = append(ageIDs, outcomes[i].Ages...)
	}
	orgs, authors, prds, sbjs, cats, sbcs, grds, ages, err := prepareAllNeededName(ctx, operator, orgIDs, authIDs, prgIDs, sbjIDs, catIDs, sbcIDs, grdIDs, ageIDs)
	if err != nil {
		log.Error(ctx, "fillOutcomeViews: prepareAllNeededName failed",
			log.Err(err),
			log.Any("outcome", outcomes))
		return nil, err
	}
	for i := range outcomes {
		outcomeViews[i] = buildOutcomeView(orgs, authors, prds, sbjs, cats, sbcs, grds, ages, outcomes[i])
	}
	return outcomeViews, nil
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

func buildOutcomeView(org, ath, prd, sbj, cat, sbc, grd, age map[string]string, outcome *entity.Outcome) *OutcomeView {
	view := &OutcomeView{
		OutcomeID:        outcome.ID,
		OutcomeName:      outcome.Name,
		AncestorID:       outcome.AncestorID,
		Shortcode:        outcome.Shortcode,
		Assumed:          outcome.Assumed,
		SourceID:         outcome.SourceID,
		LatestID:         outcome.LatestID,
		LockedBy:         outcome.LockedBy,
		AuthorID:         outcome.AuthorID,
		OrganizationID:   outcome.OrganizationID,
		AuthorName:       ath[outcome.AuthorID],
		OrganizationName: org[outcome.OrganizationID],
		PublishScope:     outcome.PublishScope,
		PublishStatus:    string(outcome.PublishStatus),
		Keywords:         strings.Split(outcome.Keywords, ","),
		RejectReason:     outcome.RejectReason,
		EstimatedTime:    outcome.EstimatedTime,
		Description:      outcome.Description,
		CreatedAt:        outcome.CreateAt,
		UpdatedAt:        outcome.UpdateAt,
	}
	view.Program = make([]Program, len(outcome.Programs))
	for k, id := range outcome.Programs {
		view.Program[k].ProgramID = id
		view.Program[k].ProgramName = prd[id]
	}

	view.Subject = make([]Subject, len(outcome.Subjects))
	for k, id := range outcome.Subjects {
		view.Subject[k].SubjectID = id
		view.Subject[k].SubjectName = sbj[id]
	}

	view.Developmental = make([]Developmental, len(outcome.Categories))
	for k, id := range outcome.Categories {
		view.Developmental[k].DevelopmentalID = id
		view.Developmental[k].DevelopmentalName = cat[id]
	}

	view.Skills = make([]Skill, len(outcome.Subcategories))
	for k, id := range outcome.Subcategories {
		view.Skills[k].SkillID = id
		view.Skills[k].SkillName = sbc[id]
	}

	view.Age = make([]Age, len(outcome.Ages))
	for k, id := range outcome.Ages {
		view.Age[k].AgeID = id
		view.Age[k].AgeName = age[id]
	}

	view.Grade = make([]Grade, len(outcome.Grades))
	for k, id := range outcome.Grades {
		view.Grade[k].GradeID = id
		view.Grade[k].GradeName = grd[id]
	}
	view.Sets = make([]*OutcomeSetCreateView, len(outcome.Sets))
	for i := range outcome.Sets {
		set := OutcomeSetCreateView{
			SetID:   outcome.Sets[i].ID,
			SetName: outcome.Sets[i].Name,
		}
		view.Sets[i] = &set
	}
	return view
}
