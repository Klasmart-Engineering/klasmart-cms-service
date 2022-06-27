package model

import (
	"context"
	"errors"
	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/da"
	"github.com/KL-Engineering/kidsloop-cms-service/da/assessmentV2"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
	"github.com/KL-Engineering/kidsloop-cms-service/utils/errgroup"
)

var (
	ErrorPreambleDataNotInitialized = errors.New("preamble data is not initialized")
)

type AssessmentListInit struct {
	ctx           context.Context
	op            *entity.Operator
	assessments   []*v2.Assessment
	assessmentMap map[string]*v2.Assessment // assessmentID

	scheduleMap         map[string]*entity.Schedule           // scheduleID
	scheduleRelationMap map[string][]*entity.ScheduleRelation // scheduleID
	assessmentUsers     []*v2.AssessmentUser

	lessPlanMap          map[string]*v2.AssessmentContentView         // lessPlanID
	teacherMap           map[string]*entity.IDName                    // userID
	programMap           map[string]*entity.IDName                    // programID
	subjectMap           map[string]*entity.IDName                    // subjectID
	classMap             map[string]*entity.IDName                    // classID
	reviewerFeedbackMap  map[string]*v2.AssessmentReviewerFeedback    // assessmentUserID
	liveRoomMap          map[string]*external.RoomInfo                // roomID
	scheduleStuReviewMap map[string]map[string]*entity.ScheduleReview // key:ScheduleID,StudentID
	assessmentUserMap    map[string][]*v2.AssessmentUser              // assessmentID
}

func NewAssessmentListInit(ctx context.Context, op *entity.Operator, assessments []*v2.Assessment) (*AssessmentListInit, error) {
	if len(assessments) <= 0 {
		return nil, constant.ErrRecordNotFound
	}

	return &AssessmentListInit{
		ctx:         ctx,
		op:          op,
		assessments: assessments,
	}, nil
}

// level 1
func (at *AssessmentListInit) initAssessmentMap() error {
	at.assessmentMap = make(map[string]*v2.Assessment, len(at.assessments))
	for _, item := range at.assessments {
		at.assessmentMap[item.ID] = item
	}

	return nil
}

//  level 2

func (at *AssessmentListInit) initScheduleMap() error {
	ctx := at.ctx
	scheduleIDs := make([]string, len(at.assessments))
	for i, item := range at.assessments {
		scheduleIDs[i] = item.ScheduleID
	}
	schedules, err := GetScheduleModel().QueryUnsafe(ctx, &entity.ScheduleQueryCondition{
		IDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
	})
	if err != nil {
		log.Error(ctx, "get schedule info error", log.Err(err), log.Strings("scheduleIDs", scheduleIDs))
		return err
	}

	result := make(map[string]*entity.Schedule, len(schedules))
	for _, item := range schedules {
		result[item.ID] = item
	}

	at.scheduleMap = result

	return nil
}

func (at *AssessmentListInit) initScheduleRelationMap() error {
	ctx := at.ctx
	op := at.op

	scheduleIDs := make([]string, len(at.assessments))
	for i, item := range at.assessments {
		scheduleIDs[i] = item.ScheduleID
	}

	scheduleRelations, err := GetScheduleRelationModel().Query(ctx, op, &da.ScheduleRelationCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
		RelationTypes: entity.NullStrings{
			Strings: []string{
				entity.ScheduleRelationTypeSubject.String(),
				entity.ScheduleRelationTypeClassRosterClass.String(),
			},
			Valid: true,
		},
	})
	if err != nil {
		return err
	}

	result := make(map[string][]*entity.ScheduleRelation, len(scheduleIDs))
	for _, item := range scheduleRelations {
		result[item.ScheduleID] = append(result[item.ScheduleID], item)
	}

	at.scheduleRelationMap = result

	return nil
}

func (at *AssessmentListInit) initScheduleReviewMap() error {
	ctx := at.ctx
	op := at.op
	result := make(map[string]map[string]*entity.ScheduleReview)

	scheduleIDs := make([]string, 0, len(at.assessments))
	for _, item := range at.assessments {
		if item.AssessmentType != v2.AssessmentTypeReviewStudy {
			continue
		}
		scheduleIDs = append(scheduleIDs, item.ScheduleID)
	}
	if len(scheduleIDs) <= 0 {
		at.scheduleStuReviewMap = result
		return nil
	}

	scheduleReviewMap, err := GetScheduleModel().GetSuccessScheduleReview(ctx, op, scheduleIDs)
	if err != nil {
		return err
	}

	for scheduleID, studentReviews := range scheduleReviewMap {
		result[scheduleID] = make(map[string]*entity.ScheduleReview)
		for _, reviewItem := range studentReviews {
			if reviewItem.LiveLessonPlan == nil {
				log.Warn(ctx, "student review content is empty", log.Any("studentReviewContent", reviewItem))
				continue
			}
			result[scheduleID][reviewItem.StudentID] = reviewItem
		}
	}

	at.scheduleStuReviewMap = result

	return nil
}

func (at *AssessmentListInit) initAssessmentUsers() error {
	ctx := at.ctx

	assessmentIDs := make([]string, len(at.assessments))
	for i, item := range at.assessments {
		assessmentIDs[i] = item.ID
	}

	var result []*v2.AssessmentUser
	err := assessmentV2.GetAssessmentUserDA().Query(ctx, &assessmentV2.AssessmentUserCondition{
		AssessmentIDs: entity.NullStrings{
			Strings: assessmentIDs,
			Valid:   true,
		},
	}, &result)
	if err != nil {
		return err
	}

	at.assessmentUsers = result

	return nil
}

// level 3

func (at *AssessmentListInit) initLessPlanMap() error {
	ctx := at.ctx

	scheduleMap := at.scheduleMap
	if scheduleMap == nil {
		log.Error(ctx, "scheduleMap data not init when get lessPlan")
		return ErrorPreambleDataNotInitialized
	}

	lockedLessPlanIDs := make([]string, 0)
	notLockedLessPlanIDs := make([]string, 0)
	for _, item := range scheduleMap {
		if item.ClassType == entity.ScheduleClassTypeHomework && (item.IsHomeFun || item.IsReview) {
			continue
		}
		if item.IsLockedLessonPlan() {
			lockedLessPlanIDs = append(lockedLessPlanIDs, item.LiveLessonPlan.LessonPlanID)
		} else {
			notLockedLessPlanIDs = append(notLockedLessPlanIDs, item.LessonPlanID)
		}
	}

	var err error
	var latestLassPlanIDMap = make(map[string]string)

	if len(notLockedLessPlanIDs) >= 0 {
		latestLassPlanIDMap, err = GetContentModel().GetLatestContentIDMapByIDListInternal(ctx, dbo.MustGetDB(ctx), notLockedLessPlanIDs)
		if err != nil {
			return err
		}

		for _, latestID := range latestLassPlanIDMap {
			lockedLessPlanIDs = append(lockedLessPlanIDs, latestID)
		}
	}

	lessPlanIDs := utils.SliceDeduplicationExcludeEmpty(lockedLessPlanIDs)
	result := make(map[string]*v2.AssessmentContentView, len(lessPlanIDs))
	if len(lessPlanIDs) <= 0 {
		at.lessPlanMap = result
		return nil
	}
	lessPlans, err := GetContentModel().GetContentByIDListInternal(ctx, dbo.MustGetDB(ctx), lessPlanIDs)
	if err != nil {
		log.Error(ctx, "get content by ids error", log.Err(err), log.Strings("lessPlanIDs", lessPlanIDs))
		return err
	}

	for _, item := range lessPlans {
		lessPlanItem := &v2.AssessmentContentView{
			ID:          item.ID,
			Name:        item.Name,
			OutcomeIDs:  item.OutcomeIDs,
			ContentType: v2.AssessmentContentTypeLessonPlan,
			LatestID:    item.LatestID,
			FileType:    item.FileType,
		}
		result[item.ID] = lessPlanItem
	}

	// update schedule lessPlan ID
	for _, item := range scheduleMap {
		if item.IsLockedLessonPlan() {
			item.LessonPlanID = item.LiveLessonPlan.LessonPlanID
		} else {
			item.LessonPlanID = latestLassPlanIDMap[item.LessonPlanID]
		}
	}

	at.lessPlanMap = result

	return nil
}

func (at *AssessmentListInit) initTeacherMap() error {
	ctx := at.ctx
	op := at.op

	assessmentUsers := at.assessmentUsers
	if assessmentUsers == nil {
		log.Error(ctx, "assessmentUsers not init when getTeacherIDs")
		return ErrorPreambleDataNotInitialized
	}

	teacherIDs := make([]string, 0)
	deDupMap := make(map[string]struct{})

	for _, auItem := range assessmentUsers {
		if auItem.UserType != v2.AssessmentUserTypeTeacher {
			continue
		}
		if _, ok := deDupMap[auItem.UserID]; !ok {
			deDupMap[auItem.UserID] = struct{}{}
			teacherIDs = append(teacherIDs, auItem.UserID)
		}
	}

	userMap, err := external.GetUserServiceProvider().BatchGetNameMap(ctx, op, teacherIDs)
	if err != nil {
		return err
	}

	result := make(map[string]*entity.IDName)
	for userID, name := range userMap {
		result[userID] = &entity.IDName{
			ID:   userID,
			Name: name,
		}
	}

	at.teacherMap = result

	return nil
}

func (at *AssessmentListInit) initProgramMap() error {
	ctx := at.ctx
	op := at.op

	scheduleMap := at.scheduleMap
	if scheduleMap == nil {
		log.Error(ctx, "scheduleMap data not init when get lessPlan")
		return ErrorPreambleDataNotInitialized
	}

	programIDs := make([]string, 0)
	deDupMap := make(map[string]struct{})
	for _, item := range scheduleMap {
		if _, ok := deDupMap[item.ProgramID]; !ok && item.ProgramID != "" {
			programIDs = append(programIDs, item.ProgramID)
			deDupMap[item.ProgramID] = struct{}{}
		}
	}

	programMap, err := external.GetProgramServiceProvider().BatchGetNameMap(ctx, op, programIDs)
	if err != nil {
		return err
	}

	result := make(map[string]*entity.IDName)
	for id, name := range programMap {
		result[id] = &entity.IDName{
			ID:   id,
			Name: name,
		}
	}

	at.programMap = result

	return nil
}

func (at *AssessmentListInit) initSubjectMap() error {
	ctx := at.ctx
	op := at.op

	relationMap := at.scheduleRelationMap
	if relationMap == nil {
		log.Error(ctx, "scheduleRelationMap data not init when get SubjectMap")
		return ErrorPreambleDataNotInitialized
	}

	subjectIDs := make([]string, 0)
	deDupMap := make(map[string]struct{})

	for _, srItems := range relationMap {
		for _, srItem := range srItems {
			if srItem.RelationType != entity.ScheduleRelationTypeSubject {
				continue
			}

			if _, ok := deDupMap[srItem.RelationID]; ok || srItem.RelationID == "" {
				continue
			}

			subjectIDs = append(subjectIDs, srItem.RelationID)
			deDupMap[srItem.RelationID] = struct{}{}
		}
	}

	subjectMap, err := external.GetSubjectServiceProvider().BatchGetNameMap(ctx, op, subjectIDs)
	if err != nil {
		return err
	}

	result := make(map[string]*entity.IDName)
	for id, name := range subjectMap {
		result[id] = &entity.IDName{
			ID:   id,
			Name: name,
		}
	}

	at.subjectMap = result

	return nil
}

func (at *AssessmentListInit) initClassMap() error {
	ctx := at.ctx
	op := at.op

	relationMap := at.scheduleRelationMap
	if relationMap == nil {
		log.Error(ctx, "scheduleRelationMap data not init when get ClassMap")
		return ErrorPreambleDataNotInitialized
	}

	classIDs := make([]string, 0)
	deDupMap := make(map[string]struct{})

	for _, srItems := range relationMap {
		for _, srItem := range srItems {
			if srItem.RelationType != entity.ScheduleRelationTypeClassRosterClass {
				continue
			}

			if _, ok := deDupMap[srItem.RelationID]; ok || srItem.RelationID == "" {
				continue
			}

			classIDs = append(classIDs, srItem.RelationID)
			deDupMap[srItem.RelationID] = struct{}{}
		}
	}

	classMap, err := external.GetClassServiceProvider().BatchGetNameMap(ctx, op, classIDs)
	if err != nil {
		return err
	}

	result := make(map[string]*entity.IDName)
	for classID, className := range classMap {
		result[classID] = &entity.IDName{
			ID:   classID,
			Name: className,
		}
	}

	at.classMap = result

	return nil
}

func (at *AssessmentListInit) initReviewerFeedbackMap() error {
	ctx := at.ctx

	assessmentMap := at.assessmentMap
	if assessmentMap == nil {
		log.Error(ctx, "assessmentMap data not init when get ReviewerFeedbackMap")
		return ErrorPreambleDataNotInitialized
	}

	assessmentUsers := at.assessmentUsers
	if assessmentUsers == nil {
		log.Error(ctx, "assessmentUsers data not init when get ReviewerFeedbackMap")
		return ErrorPreambleDataNotInitialized
	}

	assessmentUserIDs := make([]string, len(assessmentUsers))
	for i, item := range assessmentUsers {
		if assessment, ok := assessmentMap[item.AssessmentID]; ok && assessment.AssessmentType == v2.AssessmentTypeOfflineStudy {
			assessmentUserIDs[i] = item.ID
		}
	}

	if len(assessmentUserIDs) <= 0 {
		at.reviewerFeedbackMap = make(map[string]*v2.AssessmentReviewerFeedback)
		return nil
	}

	condition := &assessmentV2.AssessmentUserResultCondition{
		AssessmentUserIDs: entity.NullStrings{
			Strings: assessmentUserIDs,
			Valid:   true,
		},
	}
	var feedbacks []*v2.AssessmentReviewerFeedback
	err := assessmentV2.GetAssessmentUserResultDA().Query(ctx, condition, &feedbacks)
	if err != nil {
		log.Error(ctx, "query reviewer feedback error", log.Err(err), log.Any("condition", condition))
		return err
	}

	result := make(map[string]*v2.AssessmentReviewerFeedback, len(feedbacks))

	for _, item := range feedbacks {
		result[item.AssessmentUserID] = item
	}

	at.reviewerFeedbackMap = result

	return nil
}

func (at *AssessmentListInit) initLiveRoomStudentsScore() error {
	ctx := at.ctx
	op := at.op

	scheduleIDs := make([]string, 0, len(at.assessments))
	for _, item := range at.assessments {
		if item.AssessmentType == v2.AssessmentTypeOnlineStudy ||
			item.AssessmentType == v2.AssessmentTypeReviewStudy {
			scheduleIDs = append(scheduleIDs, item.ScheduleID)
		}
	}

	if len(scheduleIDs) <= 0 {
		at.liveRoomMap = make(map[string]*external.RoomInfo)
		return nil
	}

	roomDataMap, err := external.GetAssessmentServiceProvider().Get(ctx, op,
		scheduleIDs,
		external.WithAssessmentGetScore(true))
	if err != nil {
		log.Warn(ctx, "external service error",
			log.Err(err), log.Strings("scheduleIDs", scheduleIDs), log.Any("op", at.op))
		at.liveRoomMap = make(map[string]*external.RoomInfo)
	} else {
		at.liveRoomMap = roomDataMap
	}

	return nil
}

func (at *AssessmentListInit) initAssessmentUserMap() error {
	ctx := at.ctx

	result := make(map[string][]*v2.AssessmentUser, len(at.assessments))

	assessmentUsers := at.assessmentUsers
	if assessmentUsers == nil {
		log.Error(ctx, "assessmentUsers data not init when get AssessmentUserMap")
		return ErrorPreambleDataNotInitialized
	}

	for _, item := range assessmentUsers {
		result[item.AssessmentID] = append(result[item.AssessmentID], item)
	}

	at.assessmentUserMap = result

	return nil
}

// assessment
// async schedule, schedule_relation,schedule_review assessment_user
// async lessPlan,teacher,program,subject,class,reviewer_feedback,live

func (at *AssessmentListInit) AsyncInitData() error {
	ctx := at.ctx
	//op := at.op

	// first
	if err := at.initAssessmentMap(); err != nil {
		return err
	}

	gSecond := new(errgroup.Group)

	// second async schedule, schedule_relation, assessment_user
	gSecond.Go(func() error {
		if err := at.initScheduleMap(); err != nil {
			return err
		}

		return nil
	})

	gSecond.Go(func() error {
		if err := at.initScheduleRelationMap(); err != nil {
			return err
		}

		return nil
	})

	gSecond.Go(func() error {
		if err := at.initScheduleReviewMap(); err != nil {
			return err
		}

		return nil
	})

	gSecond.Go(func() error {
		if err := at.initAssessmentUsers(); err != nil {
			return err
		}

		return nil
	})

	if err := gSecond.Wait(); err != nil {
		log.Error(ctx, "get assessment second level info error",
			log.Err(err))
		return err
	}

	// third async lessPlan,teacher,program,subject,class,reviewer_feedback,live
	gThird := new(errgroup.Group)

	gThird.Go(func() error {
		if err := at.initLessPlanMap(); err != nil {
			return err
		}

		return nil
	})

	gThird.Go(func() error {
		if err := at.initTeacherMap(); err != nil {
			return err
		}

		return nil
	})

	gThird.Go(func() error {
		if err := at.initProgramMap(); err != nil {
			return err
		}

		return nil
	})

	gThird.Go(func() error {
		if err := at.initSubjectMap(); err != nil {
			return err
		}

		return nil
	})

	gThird.Go(func() error {
		if err := at.initClassMap(); err != nil {
			return err
		}

		return nil
	})

	gThird.Go(func() error {
		if err := at.initReviewerFeedbackMap(); err != nil {
			return err
		}

		return nil
	})

	gThird.Go(func() error {
		if err := at.initLiveRoomStudentsScore(); err != nil {
			return err
		}

		return nil
	})

	gThird.Go(func() error {
		if err := at.initAssessmentUserMap(); err != nil {
			return err
		}

		return nil
	})

	if err := gThird.Wait(); err != nil {
		log.Error(ctx, "get assessment third level info error",
			log.Err(err))
		return err
	}

	return nil
}

func (at *AssessmentListInit) MatchProgram() map[string]*entity.IDName {
	scheduleMap := at.scheduleMap
	programMap := at.programMap

	result := make(map[string]*entity.IDName, len(at.assessments))
	for _, item := range at.assessments {
		if schedule, ok := scheduleMap[item.ScheduleID]; ok {
			result[item.ID] = programMap[schedule.ProgramID]
		}
	}

	return result
}

func (at *AssessmentListInit) MatchSubject() map[string][]*entity.IDName {
	relationMap := at.scheduleRelationMap

	subjectMap := at.subjectMap

	result := make(map[string][]*entity.IDName, len(at.assessments))
	for _, item := range at.assessments {
		if srItems, ok := relationMap[item.ScheduleID]; ok {
			for _, srItem := range srItems {
				if srItem.RelationType != entity.ScheduleRelationTypeSubject {
					continue
				}
				if subItem, ok := subjectMap[srItem.RelationID]; ok && subItem != nil {
					result[item.ID] = append(result[item.ID], subItem)
				}
			}
		}
	}

	return result
}

func (at *AssessmentListInit) MatchClass() map[string]*entity.IDName {
	relationMap := at.scheduleRelationMap
	classMap := at.classMap

	result := make(map[string]*entity.IDName, len(at.assessments))
	for _, item := range at.assessments {
		if srItems, ok := relationMap[item.ScheduleID]; ok {
			for _, srItem := range srItems {
				if srItem.RelationType == entity.ScheduleRelationTypeClassRosterClass {
					result[item.ID] = classMap[srItem.RelationID]
					break
				}
			}
		}
	}

	return result
}

func ConvertAssessmentPageReply(ctx context.Context, op *entity.Operator, assessments []*v2.Assessment) ([]*v2.AssessmentQueryReply, error) {
	listInit, err := NewAssessmentListInit(ctx, op, assessments)
	if err != nil {
		return nil, err
	}

	if err = listInit.AsyncInitData(); err != nil {
		return nil, err
	}

	processorMap := make(map[v2.AssessmentType]IAssessmentProcessor)
	processorMap[v2.AssessmentTypeOnlineClass] = NewOnlineClassAssessment()
	processorMap[v2.AssessmentTypeOfflineClass] = NewOnlineClassAssessment()
	processorMap[v2.AssessmentTypeOnlineStudy] = NewOnlineStudyAssessment()
	processorMap[v2.AssessmentTypeReviewStudy] = NewReviewStudyAssessment()
	processorMap[v2.AssessmentTypeOfflineStudy] = NewOfflineStudyAssessment()

	scheduleMap := listInit.scheduleMap
	lessPlanMap := listInit.lessPlanMap
	programMap := listInit.MatchProgram()
	subjectMap := listInit.MatchSubject()
	classMap := listInit.MatchClass()

	// complete rate
	completeRateMap := listInit.MatchCompleteRate(processorMap)

	// teacherName
	teacherNameMap := listInit.MatchTeacherName(processorMap)

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

		replyItem.Teachers = teacherNameMap[item.ID]

		schedule, ok := scheduleMap[item.ScheduleID]
		if !ok {
			continue
		}

		if lessPlanItem, ok := lessPlanMap[schedule.LessonPlanID]; ok {
			replyItem.LessonPlan = &entity.IDName{
				ID:   lessPlanItem.ID,
				Name: lessPlanItem.Name,
			}
		}

		replyItem.Program = programMap[item.ID]
		replyItem.Subjects = subjectMap[item.ID]
		replyItem.DueAt = schedule.DueAt
		replyItem.ClassInfo = classMap[item.ID]
		replyItem.CompleteRate = completeRateMap[item.ID]
	}

	return result, nil
}

func (at *AssessmentListInit) MatchTeacherName(processorMap map[v2.AssessmentType]IAssessmentProcessor) map[string][]*entity.IDName {
	assessmentUserMap := at.assessmentUserMap
	teacherMap := at.teacherMap

	result := make(map[string][]*entity.IDName, len(at.assessmentMap))
	for _, item := range at.assessmentMap {
		processor := processorMap[item.AssessmentType]

		if assUserItems, ok := assessmentUserMap[item.ID]; ok {
			for _, assUserItem := range assUserItems {
				if resultItem, ok := processor.ProcessTeacherName(assUserItem, teacherMap); ok {
					result[item.ID] = append(result[item.ID], resultItem)
				}
			}
		}
	}

	return result
}

func (at *AssessmentListInit) MatchCompleteRate(processorMap map[v2.AssessmentType]IAssessmentProcessor) map[string]float64 {
	ctx := at.ctx

	assessmentUserMap := at.assessmentUserMap
	roomDataMap := at.liveRoomMap
	scheduleReviewMap := at.scheduleStuReviewMap
	reviewerFeedbackMap := at.reviewerFeedbackMap

	result := make(map[string]float64)

	for _, item := range at.assessmentMap {
		assessmentUsers := assessmentUserMap[item.ID]
		if roomData, ok := roomDataMap[item.ScheduleID]; ok {
			processor := processorMap[item.AssessmentType]

			var stuReviewMap map[string]*entity.ScheduleReview
			if scheduleReviewMap != nil {
				stuReviewMap = scheduleReviewMap[item.ScheduleID]
			}
			rate := processor.ProcessCompleteRate(ctx, assessmentUsers, roomData, stuReviewMap, reviewerFeedbackMap)
			result[item.ID] = rate
		}
	}

	return result
}

//func GetAssessmentMatch2(assessmentType v2.AssessmentType, at *AssessmentTool, action AssessmentMatchAction) IAssessmentMatch {
//	var match IAssessmentMatch
//	switch assessmentType {
//	case v2.AssessmentTypeOnlineClass:
//		match = NewOnlineClassAssessment(at, action)
//	case v2.AssessmentTypeOfflineClass:
//		match = NewOfflineClassAssessment(at, action)
//	case v2.AssessmentTypeOnlineStudy:
//		match = NewOnlineStudyAssessment(at, action)
//	case v2.AssessmentTypeReviewStudy:
//		match = NewReviewStudyAssessment(at, action)
//	case v2.AssessmentTypeOfflineStudy:
//		match = NewOfflineStudyAssessment(at, action)
//	default:
//		match = NewEmptyAssessment()
//	}
//
//	return match
//}
