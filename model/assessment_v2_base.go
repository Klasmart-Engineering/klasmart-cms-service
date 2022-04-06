package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
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
}

type AssessmentMatchAction int

const (
	_ AssessmentMatchAction = iota
	AssessmentMatchActionPage
	AssessmentMatchActionDetail
)

func NewBaseAssessment(ag *AssessmentGrain) BaseAssessment {
	return BaseAssessment{
		ag: ag,
	}
}

//func NewEmptyAssessmentDetail(ctx context.Context, op *entity.Operator, assessment *v2.Assessment) IAssessmentMatch {
//	return &BaseAssessment{
//		ctx: ctx,
//		op:  op,
//		ag: NewAssessmentsGrain(ctx, op, []*v2.Assessment{assessment}),
//	}
//}

type BaseAssessment struct {
	ag *AssessmentGrain
}

func (o *BaseAssessment) MatchAnyOneAttempted() (bool, error) {
	roomDataMap, err := o.ag.GetRoomStudentScoresAndComments()
	if err != nil {
		return false, err
	}
	roomData, ok := roomDataMap[o.ag.assessment.ScheduleID]
	return ok && roomData != nil && len(roomData.ScoresByUser) > 0, nil
}

func (o *BaseAssessment) MatchClass() (map[string]*entity.IDName, error) {
	relationMap, err := o.ag.GetScheduleRelationMap()
	if err != nil {
		return nil, err
	}

	classMap, err := o.ag.GetClassMap()
	if err != nil {
		return nil, err
	}

	result := make(map[string]*entity.IDName, len(o.ag.assessments))
	for _, item := range o.ag.assessments {
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
	scheduleMap, err := o.ag.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	return scheduleMap, nil
}

func (o *BaseAssessment) MatchLessPlan() (map[string]*v2.AssessmentContentView, error) {
	scheduleMap, err := o.ag.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	lessPlanMap, err := o.ag.GetLessPlanMap()
	if err != nil {
		return nil, err
	}

	// key:assessmentID
	result := make(map[string]*v2.AssessmentContentView, len(o.ag.assessments))
	for _, item := range o.ag.assessments {
		if schedule, ok := scheduleMap[item.ScheduleID]; ok {
			if lessPlanItem, ok := lessPlanMap[schedule.LessonPlanID]; ok && lessPlanItem != nil {
				result[item.ID] = lessPlanItem
			}
		}
	}

	return result, nil
}

func (o *BaseAssessment) MatchTeacher() (map[string][]*entity.IDName, error) {
	assessmentUserMap, err := o.ag.GetAssessmentUserMap()
	if err != nil {
		return nil, err
	}

	userMap, err := o.ag.GetUserMap()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]*entity.IDName, len(o.ag.assessments))
	for _, item := range o.ag.assessments {
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

func (o *BaseAssessment) summaryRoomScores(userScoreMap map[string][]*RoomUserScore, contentsReply []*v2.AssessmentContentReply) (map[string]float64, map[string]float64) {
	contentSummaryTotalScoreMap := make(map[string]float64)
	contentMap := make(map[string]*v2.AssessmentContentReply)
	for _, content := range contentsReply {
		if content.IgnoreCalculateScore {
			continue
		}
		contentID := content.ContentID
		if content.ContentType == v2.AssessmentContentTypeUnknown {
			contentID = content.ParentID
		}
		contentSummaryTotalScoreMap[contentID] = contentSummaryTotalScoreMap[contentID] + content.MaxScore

		contentMap[content.ContentID] = content
	}

	roomUserResultMap := make(map[string]*RoomUserScore)
	roomUserSummaryScoreMap := make(map[string]float64)
	for userID, scores := range userScoreMap {
		for _, resultItem := range scores {
			key := o.ag.GetKey([]string{
				userID,
				resultItem.ContentUniqueID,
			})
			roomUserResultMap[key] = resultItem

			if contentItem, ok := contentMap[resultItem.ContentUniqueID]; ok {
				if contentItem.IgnoreCalculateScore {
					continue
				}
				contentID := contentItem.ContentID
				if contentItem.ContentType == v2.AssessmentContentTypeUnknown {
					contentID = contentItem.ParentID
				}

				key2 := o.ag.GetKey([]string{
					userID,
					contentID,
				})
				roomUserSummaryScoreMap[key2] = roomUserSummaryScoreMap[key2] + resultItem.Score
			}
		}
	}

	log.Debug(o.ag.ctx, "summary score info", log.Any("contentSummaryTotalScoreMap", contentSummaryTotalScoreMap), log.Any("roomUserSummaryScoreMap", roomUserSummaryScoreMap))
	return contentSummaryTotalScoreMap, roomUserSummaryScoreMap
}

func GetAssessmentPageMatch(assessmentType v2.AssessmentType, ags *AssessmentGrain) IAssessmentMatch {
	var match IAssessmentMatch
	switch assessmentType {
	case v2.AssessmentTypeOnlineClass:
		match = NewOnlineClassAssessmentPage(ags)
	case v2.AssessmentTypeOfflineClass:
		match = NewOfflineClassAssessmentPage(ags)
	case v2.AssessmentTypeOnlineStudy:
		match = NewOnlineStudyAssessmentPage(ags)
	case v2.AssessmentTypeReviewStudy:
		match = NewReviewStudyAssessmentPage(ags)
	default:
		match = NewEmptyAssessment()
	}

	return match
}

func GetAssessmentDetailMatch(assessmentType v2.AssessmentType, ags *AssessmentGrain) IAssessmentMatch {
	var match IAssessmentMatch
	switch assessmentType {
	case v2.AssessmentTypeOnlineClass:
		match = NewOnlineClassAssessmentDetail(ags)
	case v2.AssessmentTypeOfflineClass:
		match = NewOfflineClassAssessmentDetail(ags)
	case v2.AssessmentTypeOnlineStudy:
		match = NewOnlineStudyAssessmentDetail(ags)
	case v2.AssessmentTypeReviewStudy:
		match = NewReviewStudyAssessmentDetail(ags)
	default:
		match = NewEmptyAssessment()
	}

	return match
}

func ConvertAssessmentPageReply(ctx context.Context, op *entity.Operator, assessmentType v2.AssessmentType, assessments []*v2.Assessment) ([]*v2.AssessmentQueryReply, error) {
	ags := NewAssessmentGrainMul(ctx, op, assessments)
	match := GetAssessmentPageMatch(assessmentType, ags)

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
			ID:         item.ID,
			Title:      item.Title,
			ClassEndAt: item.ClassEndAt,
			CompleteAt: item.CompleteAt,
			Status:     item.Status,
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
	ags := NewAssessmentGrainSingle(ctx, op, assessment)
	match := GetAssessmentDetailMatch(assessment.AssessmentType, ags)

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
		ID:           assessment.ID,
		Title:        assessment.Title,
		Status:       assessment.Status,
		RoomID:       assessment.ScheduleID,
		ClassEndAt:   assessment.ClassEndAt,
		ClassLength:  assessment.ClassLength,
		CompleteAt:   assessment.CompleteAt,
		CompleteRate: 0,
	}

	schedule, ok := scheduleMap[assessment.ScheduleID]
	if !ok {
		return nil, constant.ErrRecordNotFound
	}

	result.Class = classMap[assessment.ID]
	result.Teachers = teacherMap[assessment.ID]
	result.Program = programMap[assessment.ID]
	result.Subjects = subjectMap[assessment.ID]
	result.ScheduleTitle = schedule.Title
	result.ScheduleDueAt = schedule.DueAt
	result.RemainingTime = remainingMap[assessment.ID]
	result.Contents = contents
	result.Students = students
	result.CompleteRate = completeRateMap[assessment.ID]
	result.IsAnyOneAttempted = isAnyOneAttempted

	for _, item := range outcomeMap {
		result.Outcomes = append(result.Outcomes, item)
	}

	result.DiffContentStudents = diffContentStudents

	return result, nil
}

func ConvertAssessmentHomePageReply(ctx context.Context, op *entity.Operator, assessmentType v2.AssessmentType, assessments []*v2.Assessment) ([]*v2.AssessmentQueryReply, error) {
	ags := NewAssessmentGrainMul(ctx, op, assessments)
	match := GetAssessmentPageMatch(assessmentType, ags)

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
	if assessmentType == v2.AssessmentTypeOnlineStudy {
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
	}

	result := make([]*v2.AssessmentQueryReply, len(assessments))

	for i, item := range assessments {
		replyItem := &v2.AssessmentQueryReply{
			ID:         item.ID,
			Title:      item.Title,
			ClassEndAt: item.ClassEndAt,
			CompleteAt: item.CompleteAt,
			Status:     item.Status,
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

func NewEmptyAssessment() IAssessmentMatch {
	return EmptyAssessment{}
}

type EmptyAssessment struct{}

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
