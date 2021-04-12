package api

import (
	"context"
	"errors"
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

func (req OutcomeCreateView) outcome() (*entity.Outcome, error) {
	outcome := entity.Outcome{
		Name:          req.OutcomeName,
		Assumed:       req.Assumed,
		EstimatedTime: req.Estimated,
		Description:   req.Description,
		Shortcode:     req.Shortcode,
	}
	outcome.Program = strings.Join(req.Program, ",")
	outcome.Subject = strings.Join(req.Subject, ",")
	outcome.Developmental = strings.Join(req.Developmental, ",")
	outcome.Skills = strings.Join(req.Skills, ",")
	outcome.Grade = strings.Join(req.Grade, ",")
	outcome.Age = strings.Join(req.Age, ",")
	outcome.Keywords = strings.Join(req.Keywords, ",")

	outcome.Sets = make([]*entity.Set, len(req.Sets))
	for i := range req.Sets {
		set := &entity.Set{
			ID: req.Sets[i].SetID,
		}
		outcome.Sets[i] = set
	}

	return &outcome, nil
}

func (req OutcomeCreateView) outcomeWithID(outcomeID string) (*entity.Outcome, error) {
	outcome, err := req.outcome()
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

func newOutcomeCreateResponse(ctx context.Context, operator *entity.Operator, createView *OutcomeCreateView, outcome *entity.Outcome) OutcomeCreateResponse {
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

func newOutcomeSearchResponse(ctx context.Context, operator *entity.Operator, total int, outcomes []*entity.Outcome) (res OutcomeSearchResponse) {
	res.Total = total
	res.List = make([]*OutcomeView, len(outcomes))
	for i := range outcomes {
		view := newOutcomeView(ctx, operator, outcomes[i])
		res.List[i] = &view
	}
	return
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

func newOutcomeView(ctx context.Context, operator *entity.Operator, outcome *entity.Outcome) OutcomeView {
	view := OutcomeView{
		OutcomeID:        outcome.ID,
		OutcomeName:      outcome.Name,
		AncestorID:       outcome.AncestorID,
		Shortcode:        outcome.Shortcode,
		Assumed:          outcome.Assumed,
		SourceID:         outcome.SourceID,
		LatestID:         outcome.LatestID,
		LockedBy:         outcome.LockedBy,
		AuthorID:         outcome.AuthorID,
		AuthorName:       outcome.AuthorName,
		OrganizationID:   outcome.OrganizationID,
		OrganizationName: getOrganizationName(ctx, operator, outcome.OrganizationID),
		PublishScope:     outcome.PublishScope,
		PublishStatus:    string(outcome.PublishStatus),
		Keywords:         strings.Split(outcome.Keywords, ","),
		RejectReason:     outcome.RejectReason,
		EstimatedTime:    outcome.EstimatedTime,
		Description:      outcome.Description,
		CreatedAt:        outcome.CreateAt,
		UpdatedAt:        outcome.UpdateAt,
	}

	author, _ := external.GetUserServiceProvider().Get(ctx, operator, outcome.AuthorID)
	view.AuthorName = author.Name
	pIDs := strings.Split(outcome.Program, ",")
	pNames, _ := external.GetProgramServiceProvider().BatchGetNameMap(ctx, operator, pIDs)
	view.Program = make([]Program, len(pIDs))
	for k, id := range pIDs {
		view.Program[k].ProgramID = id
		view.Program[k].ProgramName = pNames[id]
	}
	sIDs := strings.Split(outcome.Subject, ",")
	sNames := getSubjectsName(ctx, operator, sIDs)
	view.Subject = make([]Subject, len(sIDs))
	for k, id := range sIDs {
		view.Subject[k].SubjectID = id
		view.Subject[k].SubjectName = sNames[id]
	}
	dIDs := strings.Split(outcome.Developmental, ",")
	categoryNames, _ := external.GetCategoryServiceProvider().BatchGetNameMap(ctx, operator, dIDs)
	view.Developmental = make([]Developmental, len(dIDs))
	for k, id := range dIDs {
		view.Developmental[k].DevelopmentalID = id
		view.Developmental[k].DevelopmentalName = categoryNames[id]
	}
	skIDs := strings.Split(outcome.Skills, ",")
	skNames := getSkillsName(ctx, operator, skIDs)
	view.Skills = make([]Skill, len(skIDs))
	for k, id := range skIDs {
		view.Skills[k].SkillID = id
		view.Skills[k].SkillName = skNames[id]
	}
	aIDs := strings.Split(outcome.Age, ",")
	aNames, _ := external.GetAgeServiceProvider().BatchGetNameMap(ctx, operator, aIDs)
	view.Age = make([]Age, len(aIDs))
	for k, id := range aIDs {
		view.Age[k].AgeID = id
		view.Age[k].AgeName = aNames[id]
	}
	gIDs := strings.Split(outcome.Grade, ",")
	gNames, _ := external.GetGradeServiceProvider().BatchGetNameMap(ctx, operator, gIDs)
	view.Grade = make([]Grade, len(gIDs))
	for k, id := range gIDs {
		view.Grade[k].GradeID = id
		view.Grade[k].GradeName = gNames[id]
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

func getSubjectsName(ctx context.Context, operator *entity.Operator, ids []string) (names map[string]string) {
	subjects, err := external.GetSubjectServiceProvider().BatchGet(ctx, operator, ids)
	if err != nil {
		log.Error(ctx, "getSubjectsName: BatchGet failed",
			log.Err(err),
			log.Strings("subjects_ids", ids))
		return nil
	}
	names = make(map[string]string, len(ids))
	for _, s := range subjects {
		names[s.ID] = s.Name
	}
	return
}

func getSkillsName(ctx context.Context, operator *entity.Operator, ids []string) (names map[string]string) {
	skills, err := external.GetSubCategoryServiceProvider().BatchGet(ctx, operator, ids)
	if err != nil {
		log.Error(ctx, "getSkillsName: BatchGet failed",
			log.Err(err),
			log.Strings("skill_ids", ids))
		return nil
	}
	names = make(map[string]string, len(ids))
	for _, s := range skills {
		names[s.ID] = s.Name
	}
	return
}
