package model

import (
	"context"

	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
)

type IAssessmentMatch interface {
	MatchSchedule() (map[string]*entity.Schedule, error)
	MatchTeacher() (map[string][]*entity.IDName, error)
	MatchLessPlan() (map[string]*v2.AssessmentContentView, error)
	MatchProgram() (map[string]*entity.IDName, error)
	MatchSubject() (map[string][]*entity.IDName, error)
	MatchClass() (map[string]*entity.IDName, error)
	MatchCompleteRate() (map[string]float64, error)
	MatchRemainingTime() (map[string]int64, error)
	MatchAnyOneAttempted() (bool, error)

	MatchOutcomes() (map[string]*v2.AssessmentOutcomeReply, error)
	MatchContents() ([]*v2.AssessmentContentReply, error)
	MatchStudents(contentsReply []*v2.AssessmentContentReply) ([]*v2.AssessmentStudentReply, error)
	MatchDiffContentStudents() ([]*v2.AssessmentDiffContentStudentsReply, error)

	Update(req *v2.AssessmentUpdateReq) error
}

type AssessmentMatchAction int

const (
	_ AssessmentMatchAction = iota
	AssessmentMatchActionPage
	AssessmentMatchActionDetail
)

func NewBaseAssessment(at *AssessmentTool, action AssessmentMatchAction) BaseAssessment {
	return BaseAssessment{
		at:     at,
		action: action,
	}
}

type BaseAssessment struct {
	at     *AssessmentTool
	action AssessmentMatchAction
}

func (o *BaseAssessment) MatchAnyOneAttempted() (bool, error) {
	roomDataMap, err := o.at.GetRoomStudentScoresAndComments()
	if err != nil {
		return false, err
	}
	roomData, ok := roomDataMap[o.at.first.ScheduleID]
	return ok && roomData != nil && len(roomData.ScoresByUser) > 0, nil
}

func (o *BaseAssessment) MatchClass() (map[string]*entity.IDName, error) {
	relationMap, err := o.at.GetScheduleRelationMap()
	if err != nil {
		return nil, err
	}

	classMap, err := o.at.GetClassMap()
	if err != nil {
		return nil, err
	}

	result := make(map[string]*entity.IDName, len(o.at.assessments))
	for _, item := range o.at.assessments {
		if srItems, ok := relationMap[item.ScheduleID]; ok {
			for _, srItem := range srItems {
				if srItem.RelationType == entity.ScheduleRelationTypeClassRosterClass {
					result[item.ID] = classMap[srItem.RelationID]
					break
				}
			}
		}
	}

	return result, nil
}

func (o *BaseAssessment) MatchSchedule() (map[string]*entity.Schedule, error) {
	scheduleMap, err := o.at.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	return scheduleMap, nil
}

func (o *BaseAssessment) MatchLessPlan() (map[string]*v2.AssessmentContentView, error) {
	scheduleMap, err := o.at.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	lessPlanMap, err := o.at.GetLessPlanMap()
	if err != nil {
		return nil, err
	}

	// key:assessmentID
	result := make(map[string]*v2.AssessmentContentView, len(o.at.assessments))
	for _, item := range o.at.assessments {
		if schedule, ok := scheduleMap[item.ScheduleID]; ok {
			if lessPlanItem, ok := lessPlanMap[schedule.LessonPlanID]; ok && lessPlanItem != nil {
				result[item.ID] = lessPlanItem
			}
		}
	}

	return result, nil
}

func (o *BaseAssessment) MatchTeacher() (map[string][]*entity.IDName, error) {
	assessmentUserMap, err := o.at.GetAssessmentUserMap()
	if err != nil {
		return nil, err
	}

	userMap := make(map[string]*entity.IDName)
	if o.action == AssessmentMatchActionPage {
		userMap, err = o.at.GetTeacherMap()
		if err != nil {
			return nil, err
		}
	}

	result := make(map[string][]*entity.IDName, len(o.at.assessments))
	for _, item := range o.at.assessments {
		if assUserItems, ok := assessmentUserMap[item.ID]; ok {
			for _, assUserItem := range assUserItems {
				if assUserItem.UserType != v2.AssessmentUserTypeTeacher {
					continue
				}
				//if assUserItem.StatusBySystem == v2.AssessmentUserStatusNotParticipate {
				//	continue
				//}
				resultItem := &entity.IDName{
					ID:   assUserItem.UserID,
					Name: "",
				}

				if userItem, ok := userMap[assUserItem.UserID]; ok && userItem != nil {
					resultItem.Name = userItem.Name
				}
				result[item.ID] = append(result[item.ID], resultItem)
			}
		}
	}

	return result, nil
}

func GetAssessmentMatch(assessmentType v2.AssessmentType, at *AssessmentTool, action AssessmentMatchAction) (IAssessmentMatch, AssessmentExternalInclude) {
	var match IAssessmentMatch
	var include AssessmentExternalInclude
	switch assessmentType {
	case v2.AssessmentTypeOnlineClass:
		match = NewOnlineClassAssessment(at, action)
		include = OnlineAndOfflineClassExternalDataInclude[action]
	case v2.AssessmentTypeOfflineClass:
		match = NewOfflineClassAssessment(at, action)
		include = OnlineAndOfflineClassExternalDataInclude[action]
	case v2.AssessmentTypeOnlineStudy:
		match = NewOnlineStudyAssessment(at, action)
		include = OnlineAndReviewStudyExternalDataInclude[action]
	case v2.AssessmentTypeReviewStudy:
		match = NewReviewStudyAssessment(at, action)
		include = OnlineAndReviewStudyExternalDataInclude[action]
	case v2.AssessmentTypeOfflineStudy:
		match = NewOfflineStudyAssessment(at, action)
		include = OfflineStudyExternalDataInclude[action]
	default:
		match = NewEmptyAssessment()
	}

	return match, include
}

func ConvertAssessmentPageReply(ctx context.Context, op *entity.Operator, assessmentType v2.AssessmentType, assessments []*v2.Assessment) ([]*v2.AssessmentQueryReply, error) {
	at, err := NewAssessmentTool(ctx, op, assessments)
	if err != nil {
		return nil, err
	}
	match, include := GetAssessmentMatch(assessmentType, at, AssessmentMatchActionPage)
	err = at.AsyncInitExternalData(include)
	if err != nil {
		return nil, err
	}

	scheduleMap, err := match.MatchSchedule()
	if err != nil {
		return nil, err
	}
	lessPlanMap, err := match.MatchLessPlan()
	if err != nil {
		return nil, err
	}
	teacherMap, err := match.MatchTeacher()
	if err != nil {
		return nil, err
	}
	programMap, err := match.MatchProgram()
	if err != nil {
		return nil, err
	}
	subjectMap, err := match.MatchSubject()
	if err != nil {
		return nil, err
	}

	var (
		classMap        map[string]*entity.IDName
		completeRateMap map[string]float64
		remainingMap    map[string]int64
	)
	classMap, err = match.MatchClass()
	if err != nil {
		return nil, err
	}
	completeRateMap, err = match.MatchCompleteRate()
	if err != nil {
		return nil, err
	}
	remainingMap, err = match.MatchRemainingTime()
	if err != nil {
		return nil, err
	}

	result := make([]*v2.AssessmentQueryReply, len(assessments))

	for i, item := range assessments {
		replyItem := &v2.AssessmentQueryReply{
			ID:             item.ID,
			AssessmentType: item.AssessmentType,
			Title:          item.Title,
			ClassEndAt:     item.ClassEndAt,
			CompleteAt:     item.CompleteAt,
			Status:         item.Status,
		}
		result[i] = replyItem

		replyItem.Teachers = teacherMap[item.ID]

		schedule, ok := scheduleMap[item.ScheduleID]
		if !ok {
			continue
		}
		if lessPlanItem, ok := lessPlanMap[item.ID]; ok {
			replyItem.LessonPlan = &entity.IDName{
				ID:   lessPlanItem.ID,
				Name: lessPlanItem.Name,
			}
		}

		replyItem.Program = programMap[item.ID]
		replyItem.Subjects = subjectMap[item.ID]
		replyItem.DueAt = schedule.DueAt
		replyItem.ClassInfo = classMap[item.ID]
		replyItem.RemainingTime = remainingMap[item.ID]
		replyItem.CompleteRate = completeRateMap[item.ID]
	}

	return result, nil
}

func ConvertAssessmentDetailReply(ctx context.Context, op *entity.Operator, assessment *v2.Assessment) (*v2.AssessmentDetailReply, error) {
	at, err := NewAssessmentTool(ctx, op, []*v2.Assessment{assessment})
	if err != nil {
		return nil, err
	}
	match, include := GetAssessmentMatch(assessment.AssessmentType, at, AssessmentMatchActionDetail)
	err = at.AsyncInitExternalData(include)
	if err != nil {
		return nil, err
	}

	scheduleMap, err := match.MatchSchedule()
	if err != nil {
		return nil, err
	}

	teacherMap, err := match.MatchTeacher()
	if err != nil {
		return nil, err
	}

	programMap, err := match.MatchProgram()
	if err != nil {
		return nil, err
	}

	subjectMap, err := match.MatchSubject()
	if err != nil {
		return nil, err
	}

	classMap, err := match.MatchClass()
	if err != nil {
		return nil, err
	}

	outcomeMap, err := match.MatchOutcomes()
	if err != nil {
		return nil, err
	}

	contents, err := match.MatchContents()
	if err != nil {
		return nil, err
	}

	students, err := match.MatchStudents(contents)
	if err != nil {
		return nil, err
	}

	completeRateMap, err := match.MatchCompleteRate()
	if err != nil {
		return nil, err
	}

	remainingMap, err := match.MatchRemainingTime()
	if err != nil {
		return nil, err
	}

	diffContentStudents, err := match.MatchDiffContentStudents()
	if err != nil {
		return nil, err
	}

	isAnyOneAttempted, err := match.MatchAnyOneAttempted()
	if err != nil {
		return nil, err
	}

	result := &v2.AssessmentDetailReply{
		ID:             assessment.ID,
		AssessmentType: assessment.AssessmentType,
		Title:          assessment.Title,
		Status:         assessment.Status,
		RoomID:         assessment.ScheduleID,
		ClassEndAt:     assessment.ClassEndAt,
		ClassLength:    assessment.ClassLength,
		CompleteAt:     assessment.CompleteAt,
		CompleteRate:   0,
	}

	schedule, ok := scheduleMap[assessment.ScheduleID]
	if !ok {
		return nil, constant.ErrRecordNotFound
	}

	result.Class = classMap[assessment.ID]

	if teachers, ok := teacherMap[assessment.ID]; ok {
		result.TeacherIDs = make([]string, len(teachers))
		for i, teacherItem := range teachers {
			result.TeacherIDs[i] = teacherItem.ID
		}
	}

	result.Program = programMap[assessment.ID]
	result.Subjects = subjectMap[assessment.ID]
	result.ScheduleTitle = schedule.Title
	result.ScheduleDueAt = schedule.DueAt
	result.RemainingTime = remainingMap[assessment.ID]
	result.Contents = contents
	result.Students = students
	result.CompleteRate = completeRateMap[assessment.ID]
	result.IsAnyOneAttempted = isAnyOneAttempted
	result.Description = schedule.Description

	for _, item := range outcomeMap {
		result.Outcomes = append(result.Outcomes, item)
	}

	result.DiffContentStudents = diffContentStudents

	return result, nil
}

func NewEmptyAssessment() IAssessmentMatch {
	return EmptyAssessment{}
}

type EmptyAssessment struct{}

func (o EmptyAssessment) Update(req *v2.AssessmentUpdateReq) error {
	return nil
}

func (o EmptyAssessment) MatchAnyOneAttempted() (bool, error) {
	return false, nil
}

//func (o EmptyAssessment) MatchAnyOneAttempted() (bool, error) {
//	return false, nil
//}

func (o EmptyAssessment) MatchDiffContentStudents() ([]*v2.AssessmentDiffContentStudentsReply, error) {
	return make([]*v2.AssessmentDiffContentStudentsReply, 0), nil
}

func (o EmptyAssessment) MatchOutcomes() (map[string]*v2.AssessmentOutcomeReply, error) {
	return make(map[string]*v2.AssessmentOutcomeReply), nil
}

func (o EmptyAssessment) MatchContents() ([]*v2.AssessmentContentReply, error) {
	return make([]*v2.AssessmentContentReply, 0), nil
}

func (o EmptyAssessment) MatchStudents(contentsReply []*v2.AssessmentContentReply) ([]*v2.AssessmentStudentReply, error) {
	return make([]*v2.AssessmentStudentReply, 0), nil
}

func (o EmptyAssessment) MatchClass() (map[string]*entity.IDName, error) {
	return make(map[string]*entity.IDName), nil
}

func (o EmptyAssessment) MatchCompleteRate() (map[string]float64, error) {
	return make(map[string]float64), nil
}

func (o EmptyAssessment) MatchRemainingTime() (map[string]int64, error) {
	return make(map[string]int64), nil
}

func (o EmptyAssessment) MatchSchedule() (map[string]*entity.Schedule, error) {
	return make(map[string]*entity.Schedule), nil
}

func (o EmptyAssessment) MatchLessPlan() (map[string]*v2.AssessmentContentView, error) {
	return make(map[string]*v2.AssessmentContentView), nil
}

func (o EmptyAssessment) MatchProgram() (map[string]*entity.IDName, error) {
	return make(map[string]*entity.IDName), nil
}

func (o EmptyAssessment) MatchSubject() (map[string][]*entity.IDName, error) {
	return make(map[string][]*entity.IDName), nil
}

func (o EmptyAssessment) MatchTeacher() (map[string][]*entity.IDName, error) {
	return make(map[string][]*entity.IDName), nil
}

var OnlineAndOfflineClassExternalDataInclude = map[AssessmentMatchAction]AssessmentExternalInclude{
	AssessmentMatchActionPage: {
		userServiceInclude: &ExternalUserServiceInclude{
			Program: true,
			Subject: true,
			Teacher: true,
		},
	},
	AssessmentMatchActionDetail: {
		userServiceInclude: &ExternalUserServiceInclude{
			Program: true,
			Subject: true,
			Teacher: true,
			Class:   true,
		},
		assessmentServiceInclude: &ExternalAssessmentServiceInclude{
			StudentScore:   true,
			TeacherComment: true,
		},
	},
}
var OnlineAndReviewStudyExternalDataInclude = map[AssessmentMatchAction]AssessmentExternalInclude{
	AssessmentMatchActionPage: {
		userServiceInclude: &ExternalUserServiceInclude{
			Teacher: true,
			Class:   true,
		},
		assessmentServiceInclude: &ExternalAssessmentServiceInclude{
			StudentScore: true,
		},
	},
	AssessmentMatchActionDetail: {
		userServiceInclude: &ExternalUserServiceInclude{
			Teacher: true,
			Class:   true,
		},
		assessmentServiceInclude: &ExternalAssessmentServiceInclude{
			StudentScore:   true,
			TeacherComment: true,
		},
	},
}

var OfflineStudyExternalDataInclude = map[AssessmentMatchAction]AssessmentExternalInclude{
	AssessmentMatchActionPage: {
		userServiceInclude: &ExternalUserServiceInclude{
			Teacher: true,
			Class:   true,
		},
	},
	AssessmentMatchActionDetail: {
		userServiceInclude: &ExternalUserServiceInclude{
			Class: true,
		},
	},
}
