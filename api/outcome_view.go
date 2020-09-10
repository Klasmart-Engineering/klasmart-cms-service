package api

import (
	"errors"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"
)

type OutcomeCreateView struct {
	OutcomeID   string `json:"outcome_id"`
	OutcomeName string `json:"outcome_name"`
	Assumed     bool   `json:"assumed"`
	//AuthorID string `json:"author_id"`
	//AuthorName string `json:"author_name"`
	//Shortcode string `json:"shortcode"`
	OrganizationID string   `json:"organization_id"`
	Program        []string `json:"program"`
	Subject        []string `json:"subject"`
	Developmental  []string `json:"developmental"`
	Skills         []string `json:"skills"`
	Age            []string `json:"age"`
	Grade          []string `json:"grade"`
	Estimated      int      `json:"estimated"`
	Keywords       []string `json:"keywords"`
	Description    string   `json:"description"`
}

func (req OutcomeCreateView) outcome() (*entity.Outcome, error) {
	outcome := entity.Outcome{
		Name:          req.OutcomeName,
		Assumed:       req.Assumed,
		EstimatedTime: req.Estimated,
		Description:   req.Description,
	}
	outcome.Program = strings.Join(req.Program, ",")
	outcome.Subject = strings.Join(req.Subject, ",")
	outcome.Developmental = strings.Join(req.Developmental, ",")
	outcome.Skills = strings.Join(req.Skills, ",")
	outcome.Grade = strings.Join(req.Grade, ",")
	outcome.Age = strings.Join(req.Age, ",")
	outcome.Keywords = strings.Join(req.Keywords, ",")
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
}

func newOutcomeCreateResponse(createView *OutcomeCreateView, outcome *entity.Outcome) OutcomeCreateResponse {
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
		OrganizationName: getProgramName(outcome.OrganizationID),
		PublishScope:     outcome.PublishScope,
		PublishStatus:    string(outcome.PublishStatus),
		RejectReason:     outcome.RejectReason,
		Description:      outcome.Description,
		CreatedAt:        outcome.CreateAt,
	}
}

type OutcomeView struct {
	OutcomeID        string          `json:"outcome_id"`
	OutcomeName      string          `json:"outcome_name"`
	AncestorID       string          `json:"ancestor_id"`
	Shortcode        string          `json:"shortcode"`
	Assumed          bool            `json:"assumed"`
	Program          []Program       `json:"program"`
	Subject          []Subject       `json:"subject"`
	Developmental    []Developmental `json:"developmental"`
	Skills           []Skill         `json:"skills"`
	Age              []Age           `json:"age"`
	Grade            []Grade         `json:"grade"`
	EstimatedTime    int             `json:"estimated_time"`
	Keywords         []string        `json:"keywords"`
	SourceID         string          `json:"source_id"`
	LockedBy         string          `json:"locked_by"`
	AuthorID         string          `json:"author_id"`
	AuthorName       string          `json:"author_name"`
	OrganizationID   string          `json:"organization_id"`
	OrganizationName string          `json:"organization_name"`
	PublishScope     string          `json:"publish_scope"`
	PublishStatus    string          `json:"publish_status"`
	RejectReason     string          `json:"reject_reason"`
	Description      string          `json:"description"`
	CreatedAt        int64           `json:"created_at"`
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

func newOutcomeView(outcome *entity.Outcome) OutcomeView {
	view := OutcomeView{
		OutcomeID:      outcome.ID,
		OutcomeName:    outcome.Name,
		AncestorID:     outcome.AncestorID,
		Shortcode:      outcome.Shortcode,
		Assumed:        outcome.Assumed,
		SourceID:       outcome.SourceID,
		LockedBy:       outcome.LockedBy,
		AuthorID:       outcome.AuthorID,
		AuthorName:     outcome.AuthorName,
		OrganizationID: outcome.OrganizationID,
		PublishScope:   outcome.PublishScope,
		PublishStatus:  string(outcome.PublishStatus),
		RejectReason:   outcome.RejectReason,
		Description:    outcome.Description,
		CreatedAt:      outcome.CreateAt,
	}
	pIDs := strings.Split(outcome.Program, ",")
	pNames := getProgramsName(pIDs)
	view.Program = make([]Program, len(pIDs))
	for k, id := range pIDs {
		view.Program[k].ProgramID = id
		view.Program[k].ProgramName = pNames[id]
	}
	sIDs := strings.Split(outcome.Subject, ",")
	sNames := getSubjectsName(sIDs)
	view.Subject = make([]Subject, len(sIDs))
	for k, id := range sIDs {
		view.Subject[k].SubjectID = id
		view.Subject[k].SubjectName = sNames[id]
	}
	dIDs := strings.Split(outcome.Developmental, ",")
	dNames := getDevelopmentalsName(dIDs)
	view.Developmental = make([]Developmental, len(dIDs))
	for k, id := range dIDs {
		view.Developmental[k].DevelopmentalID = id
		view.Developmental[k].DevelopmentalName = dNames[id]
	}
	skIDs := strings.Split(outcome.Skills, ",")
	skNames := getSkillsName(skIDs)
	view.Skills = make([]Skill, len(skIDs))
	for k, id := range skIDs {
		view.Skills[k].SkillID = id
		view.Skills[k].SkillName = skNames[id]
	}
	aIDs := strings.Split(outcome.Age, ",")
	aNames := getAgesName(aIDs)
	view.Age = make([]Age, len(aIDs))
	for k, id := range aIDs {
		view.Age[k].AgeID = id
		view.Age[k].AgeName = aNames[id]
	}
	gIDs := strings.Split(outcome.Grade, ",")
	gNames := getGradeName(gIDs)
	view.Grade = make([]Grade, len(gIDs))
	for k, id := range gIDs {
		view.Grade[k].GradeID = id
		view.Grade[k].GradeName = gNames[id]
	}
	return view
}

func getProgramName(id string) (name string) {
	// TODO : now is mock
	name = "mock Program" + id
	return
}
func getProgramsName(IDs []string) (names map[string]string) {
	// TODO : now is mock
	names = make(map[string]string, len(IDs))
	for _, id := range IDs {
		names[id] = "mock Program" + id
	}
	return
}

func getSubjectsName(IDs []string) (names map[string]string) {
	// TODO : now is mock
	names = make(map[string]string, len(IDs))
	for _, id := range IDs {
		names[id] = "mock Subject" + id
	}
	return
}

func getDevelopmentalsName(IDs []string) (names map[string]string) {
	// TODO : now is mock
	names = make(map[string]string, len(IDs))
	for _, id := range IDs {
		names[id] = "mock Developmental" + id
	}
	return
}

func getSkillsName(IDs []string) (names map[string]string) {
	// TODO : now is mock
	names = make(map[string]string, len(IDs))
	for _, id := range IDs {
		names[id] = "mock Skill" + id
	}
	return
}

func getAgesName(IDs []string) (names map[string]string) {
	// TODO : now is mock
	names = make(map[string]string, len(IDs))
	for _, id := range IDs {
		names[id] = "mock Age" + id
	}
	return
}

func getGradeName(IDs []string) (names map[string]string) {
	// TODO : now is mock
	names = make(map[string]string, len(IDs))
	for _, id := range IDs {
		names[id] = "mock Grade" + id
	}
	return
}
