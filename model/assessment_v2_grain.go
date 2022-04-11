package model

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da/assessmentV2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"strings"
)

type AssessmentsGrainInit int

const (
	_ AssessmentsGrainInit = iota
	GrainAssessmentMap
	GrainScheduleRelationMap
	GrainAssessmentUserMap
	GrainAssessmentContentMap
	GrainAssessmentUserSlice

	GrainScheduleMap
	GrainProgramMap
	GrainSubjectMap
	GrainClassMap
	GrainUserMap
	GrainLiveRoomMap
	GrainLessPlanMap
	GrainAssessmentReviewerFeedbackMap

	SingleOutcomeMapFromContent
	SingleLatestContentSliceFromSchedule
	SingleLockedContentSliceFromSchedule
	SingleContentMapFromAssessment
	SingleContentMapFromLiveRoom
	SingleUserCommentResultMap
	SingleOutcomeMapFromAssessment
	SingleGetOutcomesFromSchedule
)

type AssessmentGrain struct {
	ctx        context.Context
	op         *entity.Operator
	initRecord map[AssessmentsGrainInit]bool

	*AssessmentMulGrainItem

	*AssessmentSingleGrainItem
}

func NewAssessmentGrainMul(ctx context.Context, op *entity.Operator, assessments []*v2.Assessment) *AssessmentGrain {
	if len(assessments) <= 0 {
		log.Warn(ctx, "assessments is empty")
		return new(AssessmentGrain)
	}

	initRecord := make(map[AssessmentsGrainInit]bool)
	firstAssessment := assessments[0]
	amg := NewAssessmentMulGrainItem(ctx, op, assessments, initRecord)
	asg := NewAssessmentSingleGrainItem(ctx, op, firstAssessment, initRecord, amg)
	return &AssessmentGrain{
		ctx:        ctx,
		op:         op,
		initRecord: initRecord,

		AssessmentMulGrainItem:    amg,
		AssessmentSingleGrainItem: asg,
	}
}

func NewAssessmentGrainSingle(ctx context.Context, op *entity.Operator, assessment *v2.Assessment) *AssessmentGrain {
	initRecord := make(map[AssessmentsGrainInit]bool)
	amg := NewAssessmentMulGrainItem(ctx, op, []*v2.Assessment{assessment}, initRecord)
	asg := NewAssessmentSingleGrainItem(ctx, op, assessment, initRecord, amg)
	return &AssessmentGrain{
		ctx:        ctx,
		op:         op,
		initRecord: initRecord,

		AssessmentMulGrainItem:    amg,
		AssessmentSingleGrainItem: asg,
	}
}

type AssessmentMulGrainItem struct {
	ctx        context.Context
	op         *entity.Operator
	initRecord map[AssessmentsGrainInit]bool

	assessments []*v2.Assessment

	assessmentMap        map[string]*v2.Assessment             // assessmentID
	scheduleMap          map[string]*entity.Schedule           // scheduleID
	scheduleRelationMap  map[string][]*entity.ScheduleRelation // scheduleID
	assessmentUserMap    map[string][]*v2.AssessmentUser       // assessmentID
	assessmentContentMap map[string][]*v2.AssessmentContent    // assessmentID
	programMap           map[string]*entity.IDName             // programID
	subjectMap           map[string]*entity.IDName             // subjectID
	classMap             map[string]*entity.IDName             // classID
	userMap              map[string]*entity.IDName             // userID
	liveRoomMap          map[string]*external.RoomInfo         // roomID
	lessPlanMap          map[string]*v2.AssessmentContentView  // lessPlanID

	assessmentReviewerFeedbackMap map[string]*v2.AssessmentReviewerFeedback // assessmentUserID

	assessmentUsers []*v2.AssessmentUser
}

func NewAssessmentMulGrainItem(ctx context.Context, op *entity.Operator, assessments []*v2.Assessment, initRecord map[AssessmentsGrainInit]bool) *AssessmentMulGrainItem {
	return &AssessmentMulGrainItem{
		ctx:         ctx,
		op:          op,
		assessments: assessments,
		initRecord:  initRecord,
	}
}

type AssessmentSingleGrainItem struct {
	ctx        context.Context
	op         *entity.Operator
	initRecord map[AssessmentsGrainInit]bool

	assessment *v2.Assessment
	amg        *AssessmentMulGrainItem

	outcomeMapFromContent      map[string]*entity.Outcome // key: outcomeID
	latestContentsFromSchedule []*v2.AssessmentContentView
	lockedContentsFromSchedule []*v2.AssessmentContentView
	contentMapFromAssessment   map[string]*v2.AssessmentContent     // key:contentID
	contentMapFromLiveRoom     map[string]*RoomContentTree          // key: contentID
	commentResultMap           map[string]string                    // userID
	outcomeMapFromAssessment   map[string]*v2.AssessmentUserOutcome // key: AssessmentUserID+AssessmentContentID+OutcomeID
	outcomesFromSchedule       []*entity.Outcome
}

func NewAssessmentSingleGrainItem(ctx context.Context, op *entity.Operator, assessment *v2.Assessment, initRecord map[AssessmentsGrainInit]bool, amg *AssessmentMulGrainItem) *AssessmentSingleGrainItem {
	return &AssessmentSingleGrainItem{
		ctx:        ctx,
		op:         op,
		assessment: assessment,
		initRecord: initRecord,
		amg:        amg,
	}
}

// Multiple assessment data processing
func (ag *AssessmentMulGrainItem) GetAssessmentMap() (map[string]*v2.Assessment, error) {
	if ag.initRecord[GrainAssessmentMap] {
		return ag.assessmentMap, nil
	}

	ag.assessmentMap = make(map[string]*v2.Assessment, len(ag.assessments))
	for _, item := range ag.assessments {
		ag.assessmentMap[item.ID] = item
	}

	ag.initRecord[GrainAssessmentMap] = true

	return ag.assessmentMap, nil
}

func (ag *AssessmentMulGrainItem) GetScheduleMap() (map[string]*entity.Schedule, error) {
	if ag.initRecord[GrainScheduleMap] {
		return ag.scheduleMap, nil
	}

	ctx := ag.ctx
	scheduleIDs := make([]string, len(ag.assessments))
	for i, item := range ag.assessments {
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
		return nil, err
	}

	ag.scheduleMap = make(map[string]*entity.Schedule, len(schedules))
	for _, item := range schedules {
		ag.scheduleMap[item.ID] = item
	}

	ag.initRecord[GrainScheduleMap] = true

	return ag.scheduleMap, nil
}

func (ag *AssessmentMulGrainItem) GetScheduleRelationMap() (map[string][]*entity.ScheduleRelation, error) {
	if ag.initRecord[GrainScheduleRelationMap] {
		return ag.scheduleRelationMap, nil
	}
	ctx := ag.ctx
	op := ag.op

	scheduleIDs := make([]string, len(ag.assessments))
	for i, item := range ag.assessments {
		scheduleIDs[i] = item.ScheduleID
	}

	scheduleRelations, err := GetScheduleRelationModel().Query(ctx, op, &da.ScheduleRelationCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
	})
	if err != nil {
		return nil, err
	}

	ag.scheduleRelationMap = make(map[string][]*entity.ScheduleRelation, len(scheduleIDs))
	for _, item := range scheduleRelations {
		ag.scheduleRelationMap[item.ScheduleID] = append(ag.scheduleRelationMap[item.ScheduleID], item)
	}

	ag.initRecord[GrainScheduleRelationMap] = true

	return ag.scheduleRelationMap, nil
}

func (ag *AssessmentMulGrainItem) GetAssessmentUsers() ([]*v2.AssessmentUser, error) {
	if ag.initRecord[GrainAssessmentUserSlice] {
		return ag.assessmentUsers, nil
	}

	ctx := ag.ctx

	assessmentIDs := make([]string, len(ag.assessments))
	for i, item := range ag.assessments {
		assessmentIDs[i] = item.ID
	}

	var assessmentUsers []*v2.AssessmentUser
	err := assessmentV2.GetAssessmentUserDA().Query(ctx, &assessmentV2.AssessmentUserCondition{
		AssessmentIDs: entity.NullStrings{
			Strings: assessmentIDs,
			Valid:   true,
		},
	}, &assessmentUsers)
	if err != nil {
		return nil, err
	}

	ag.assessmentUsers = assessmentUsers
	ag.initRecord[GrainAssessmentUserSlice] = true

	return ag.assessmentUsers, nil
}

func (ag *AssessmentMulGrainItem) GetAssessmentUserMap() (map[string][]*v2.AssessmentUser, error) {
	if ag.initRecord[GrainAssessmentUserMap] {
		return ag.assessmentUserMap, nil
	}

	assessmentUsers, err := ag.GetAssessmentUsers()
	if err != nil {
		return nil, err
	}

	ag.assessmentUserMap = make(map[string][]*v2.AssessmentUser, len(ag.assessments))
	for _, item := range assessmentUsers {
		ag.assessmentUserMap[item.AssessmentID] = append(ag.assessmentUserMap[item.AssessmentID], item)
	}

	ag.initRecord[GrainAssessmentUserMap] = true

	return ag.assessmentUserMap, nil
}

func (ag *AssessmentMulGrainItem) GetProgramMap() (map[string]*entity.IDName, error) {
	if ag.initRecord[GrainProgramMap] {
		return ag.programMap, nil
	}

	ctx := ag.ctx
	op := ag.op

	scheduleMap, err := ag.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	programIDs := make([]string, 0)
	deDupMap := make(map[string]struct{})
	for _, item := range ag.assessments {
		schedule, ok := scheduleMap[item.ScheduleID]
		if !ok {
			log.Warn(ctx, "schedule not found", log.String("scheduleID", item.ScheduleID))
			continue
		}

		if _, ok := deDupMap[schedule.ProgramID]; !ok && schedule.ProgramID != "" {
			programIDs = append(programIDs, schedule.ProgramID)
			deDupMap[schedule.ProgramID] = struct{}{}
		}
	}

	programs, err := external.GetProgramServiceProvider().BatchGet(ctx, op, programIDs)
	if err != nil {
		return nil, err
	}

	ag.programMap = make(map[string]*entity.IDName)
	for _, item := range programs {
		if item == nil || item.ID == "" {
			log.Warn(ctx, "program id is empty", log.Any("programs", programs))
			continue
		}
		ag.programMap[item.ID] = &entity.IDName{
			ID:   item.ID,
			Name: item.Name,
		}
	}

	ag.initRecord[GrainProgramMap] = true

	return ag.programMap, nil
}

func (ag *AssessmentMulGrainItem) GetSubjectMap() (map[string]*entity.IDName, error) {
	if ag.initRecord[GrainSubjectMap] {
		return ag.subjectMap, nil
	}

	ctx := ag.ctx
	op := ag.op

	relationMap, err := ag.GetScheduleRelationMap()
	if err != nil {
		return nil, err
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

	subjects, err := external.GetSubjectServiceProvider().BatchGet(ctx, op, subjectIDs)
	if err != nil {
		return nil, err
	}

	ag.subjectMap = make(map[string]*entity.IDName)
	for _, item := range subjects {
		if item == nil || item.ID == "" {
			log.Warn(ctx, "subject is empty", log.Any("subjects", subjects))
			continue
		}
		ag.subjectMap[item.ID] = &entity.IDName{
			ID:   item.ID,
			Name: item.Name,
		}
	}

	ag.initRecord[GrainSubjectMap] = true

	return ag.subjectMap, nil
}

func (ag *AssessmentMulGrainItem) GetClassMap() (map[string]*entity.IDName, error) {
	if ag.initRecord[GrainClassMap] {
		return ag.classMap, nil
	}

	ctx := ag.ctx
	op := ag.op

	relationMap, err := ag.GetScheduleRelationMap()
	if err != nil {
		return nil, err
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

	classes, err := external.GetClassServiceProvider().BatchGet(ctx, op, classIDs)
	if err != nil {
		return nil, err
	}

	ag.classMap = make(map[string]*entity.IDName)
	for _, item := range classes {
		if item == nil || item.ID == "" {
			continue
		}
		ag.classMap[item.ID] = &entity.IDName{
			ID:   item.ID,
			Name: item.Name,
		}
	}

	ag.initRecord[GrainClassMap] = true

	return ag.classMap, nil
}

func (ag *AssessmentMulGrainItem) GetUserMap() (map[string]*entity.IDName, error) {
	if ag.initRecord[GrainUserMap] {
		return ag.userMap, nil
	}

	ctx := ag.ctx
	op := ag.op

	assessmentUsers, err := ag.GetAssessmentUsers()
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(assessmentUsers))
	deDupMap := make(map[string]struct{})

	for _, auItem := range assessmentUsers {
		if _, ok := deDupMap[auItem.UserID]; !ok {
			deDupMap[auItem.UserID] = struct{}{}
			userIDs = append(userIDs, auItem.UserID)
		}
	}

	users, err := external.GetUserServiceProvider().BatchGet(ctx, op, userIDs)
	if err != nil {
		return nil, err
	}

	ag.userMap = make(map[string]*entity.IDName)
	for _, item := range users {
		if !item.Valid {
			log.Warn(ctx, "user is inValid", log.Any("item", item))
			continue
		}
		ag.userMap[item.ID] = &entity.IDName{
			ID:   item.ID,
			Name: item.Name,
		}
	}

	ag.initRecord[GrainUserMap] = true

	return ag.userMap, nil
}

func (ag *AssessmentMulGrainItem) GetRoomStudentScoresAndComments() (map[string]*external.RoomInfo, error) {
	if ag.initRecord[GrainLiveRoomMap] {
		return ag.liveRoomMap, nil
	}

	ctx := ag.ctx
	op := ag.op

	scheduleMap, err := ag.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	scheduleIDs := make([]string, 0, len(ag.assessments))
	for _, item := range scheduleMap {
		scheduleIDs = append(scheduleIDs, item.ID)
	}

	roomDataMap, err := external.GetAssessmentServiceProvider().GetScoresWithCommentsByRoomIDs(ctx, op, scheduleIDs)
	if err != nil {
		log.Warn(ctx, "external service error",
			log.Err(err), log.Strings("scheduleIDs", scheduleIDs), log.Any("op", ag.op))
		ag.liveRoomMap = make(map[string]*external.RoomInfo)
	} else {
		ag.liveRoomMap = roomDataMap
	}

	ag.initRecord[GrainLiveRoomMap] = true

	return ag.liveRoomMap, nil
}

func (ag *AssessmentMulGrainItem) GetLessPlanMap() (map[string]*v2.AssessmentContentView, error) {
	ctx := ag.ctx

	if ag.initRecord[GrainLessPlanMap] {
		return ag.lessPlanMap, nil
	}

	scheduleMap, err := ag.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	lockedLessPlanIDs := make([]string, 0)
	notLockedLessPlanIDs := make([]string, 0)
	for _, item := range scheduleMap {
		if item.IsLockedLessonPlan() {
			lockedLessPlanIDs = append(lockedLessPlanIDs, item.LiveLessonPlan.LessonPlanID)
		} else {
			notLockedLessPlanIDs = append(notLockedLessPlanIDs, item.LessonPlanID)
		}
	}

	latestLassPlanIDMap, err := GetContentModel().GetLatestContentIDMapByIDListInternal(ctx, dbo.MustGetDB(ctx), notLockedLessPlanIDs)
	if err != nil {
		return nil, err
	}

	for _, latestID := range latestLassPlanIDMap {
		lockedLessPlanIDs = append(lockedLessPlanIDs, latestID)
	}
	lessPlanIDs := utils.SliceDeduplicationExcludeEmpty(lockedLessPlanIDs)
	lessPlans, err := GetContentModel().GetContentByIDListInternal(ctx, dbo.MustGetDB(ctx), lessPlanIDs)
	if err != nil {
		log.Error(ctx, "get content by ids error", log.Err(err), log.Strings("lessPlanIDs", lessPlanIDs))
		return nil, err
	}

	ag.lessPlanMap = make(map[string]*v2.AssessmentContentView, len(lessPlans))
	for _, item := range lessPlans {
		lessPlanItem := &v2.AssessmentContentView{
			ID:          item.ID,
			Name:        item.Name,
			OutcomeIDs:  item.OutcomeIDs,
			ContentType: v2.AssessmentContentTypeLessonPlan,
			LatestID:    item.LatestID,
			FileType:    item.FileType,
		}
		ag.lessPlanMap[item.ID] = lessPlanItem
	}

	// update schedule lessPlan ID
	for _, item := range scheduleMap {
		if item.IsLockedLessonPlan() {
			item.LessonPlanID = item.LiveLessonPlan.LessonPlanID
		} else {
			item.LessonPlanID = latestLassPlanIDMap[item.LessonPlanID]
		}
	}

	ag.initRecord[GrainLessPlanMap] = true

	return ag.lessPlanMap, nil
}

func (ag *AssessmentMulGrainItem) GetReviewerFeedbackMap() (map[string]*v2.AssessmentReviewerFeedback, error) {
	if ag.initRecord[GrainAssessmentReviewerFeedbackMap] {
		return ag.assessmentReviewerFeedbackMap, nil
	}

	ctx := ag.ctx

	result := make(map[string]*v2.AssessmentReviewerFeedback)

	assessmentUsers, err := ag.GetAssessmentUsers()
	if err != nil {
		return nil, err
	}

	assessmentUserIDs := make([]string, len(assessmentUsers))
	for i, item := range assessmentUsers {
		assessmentUserIDs[i] = item.ID
	}

	condition := &assessmentV2.AssessmentUserResultCondition{
		AssessmentUserIDs: entity.NullStrings{
			Strings: assessmentUserIDs,
			Valid:   true,
		},
	}
	var feedbacks []*v2.AssessmentReviewerFeedback
	err = assessmentV2.GetAssessmentUserResultDA().Query(ctx, condition, &feedbacks)
	if err != nil {
		log.Error(ctx, "query reviewer feedback error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}

	for _, item := range feedbacks {
		result[item.AssessmentUserID] = item
	}

	ag.assessmentReviewerFeedbackMap = result
	ag.initRecord[GrainAssessmentReviewerFeedbackMap] = true

	return result, nil
}

// Single assessment data processing
func (asg *AssessmentSingleGrainItem) GetAssessmentUserWithUserIDAndUserTypeMap() (map[string]*v2.AssessmentUser, error) {
	ctx := asg.ctx

	assessmentUserMap, err := asg.amg.GetAssessmentUserMap()
	if err != nil {
		return nil, err
	}

	assessmentUsers, ok := assessmentUserMap[asg.assessment.ID]
	if !ok {
		log.Error(ctx, "not found assessment users", log.Any("assessment", asg.assessment), log.Any("assessmentUserMap", assessmentUserMap))
		return nil, constant.ErrRecordNotFound
	}

	result := make(map[string]*v2.AssessmentUser, len(assessmentUsers))
	for _, item := range assessmentUsers {
		result[asg.GetKey([]string{item.UserID, item.UserType.String()})] = item
	}

	return result, nil
}

func (asg *AssessmentSingleGrainItem) GetKey(value []string) string {
	return strings.Join(value, "_")
}

func (asg *AssessmentSingleGrainItem) IsNeedConvertLatestContent() (bool, error) {
	ctx := asg.ctx

	schedule, err := asg.SingleGetSchedule()
	if err != nil {
		return false, err
	}

	if (asg.assessment.MigrateFlag == constant.AssessmentHistoryFlag &&
		asg.assessment.Status != v2.AssessmentStatusNotStarted) || !schedule.IsLockedLessonPlan() {
		log.Debug(ctx, "assessment belongs to the migration or schedule can not locked lessPlan", log.Any("assessment", asg.assessment))
		return true, nil
	}

	return false, nil
}

func (asg *AssessmentSingleGrainItem) ConvertContentOutcome(contents []*v2.AssessmentContentView) error {
	ctx := asg.ctx
	op := asg.op

	contentOutcomes, err := assessmentV2.GetAssessmentUserOutcomeDA().GetContentOutcomeByAssessmentID(ctx, asg.assessment.ID)
	if err != nil {
		return err
	}

	if len(contentOutcomes) <= 0 {
		outcomeIDs := make([]string, 0)
		dedup := make(map[string]struct{})
		for _, contentItem := range contents {
			for _, outcomeID := range contentItem.OutcomeIDs {
				if outcomeID == "" {
					continue
				}

				if _, ok := dedup[outcomeID]; !ok {
					outcomeIDs = append(outcomeIDs, outcomeID)
				}
			}
		}

		latestOutcomeMap, _, err := GetOutcomeModel().GetLatestOutcomes(ctx, op, dbo.MustGetDB(ctx), outcomeIDs)
		if err != nil {
			return err
		}

		for _, contentItem := range contents {
			newOutcomeIDMap := make(map[string]struct{})
			for _, outcomeID := range contentItem.OutcomeIDs {
				if latestOutcomeItem, ok := latestOutcomeMap[outcomeID]; ok {
					newOutcomeIDMap[latestOutcomeItem.ID] = struct{}{}
				}
			}

			contentItem.OutcomeIDs = make([]string, 0, len(newOutcomeIDMap))
			for key, _ := range newOutcomeIDMap {
				contentItem.OutcomeIDs = append(contentItem.OutcomeIDs, key)
			}
		}
	} else {
		contentOutcomeMap := make(map[string][]string)
		for _, item := range contentOutcomes {
			if item.OutcomeID != "" {
				contentOutcomeMap[item.ContentID] = append(contentOutcomeMap[item.ContentID], item.OutcomeID)
			}
		}

		for _, contentItem := range contents {
			if outcomeIDs, ok := contentOutcomeMap[contentItem.ID]; ok {
				contentItem.OutcomeIDs = outcomeIDs
			}
		}
	}

	return nil
}

func (asg *AssessmentSingleGrainItem) SingleGetLatestContentsFromSchedule() ([]*v2.AssessmentContentView, error) {
	if asg.initRecord[SingleLatestContentSliceFromSchedule] {
		return asg.latestContentsFromSchedule, nil
	}

	ctx := asg.ctx
	op := asg.op

	schedule, err := asg.SingleGetSchedule()
	if err != nil {
		return nil, err
	}

	// convert to latest lesson plan
	latestLessPlanIDMap, err := GetContentModel().GetLatestContentIDMapByIDListInternal(ctx, dbo.MustGetDB(ctx), []string{schedule.LessonPlanID})
	if err != nil {
		return nil, err
	}
	latestLessPlanID, ok := latestLessPlanIDMap[schedule.LessonPlanID]
	if !ok {
		log.Error(ctx, "lessPlan not found", log.Any("schedule", schedule), log.Any("latestLessPlanIDMap", latestLessPlanIDMap))
		return nil, constant.ErrRecordNotFound
	}

	// get lesson plan info
	latestLessPlans, err := GetContentModel().GetContentByIDListInternal(ctx, dbo.MustGetDB(ctx), []string{latestLessPlanID})
	if err != nil {
		return nil, err
	}
	if len(latestLessPlans) <= 0 {
		log.Warn(ctx, "not found content info", log.String("latestLessPlanID", latestLessPlanID), log.Any("schedule", schedule))
		return nil, constant.ErrRecordNotFound
	}

	// get material in lesson plan
	subContentsMap, err := GetContentModel().GetContentsSubContentsMapByIDListInternal(ctx, dbo.MustGetDB(ctx), []string{latestLessPlanID}, op)
	if err != nil {
		return nil, err
	}

	// filling lesson plan
	latestLessPlan := latestLessPlans[0]
	lessPlan := &v2.AssessmentContentView{
		ID:          latestLessPlan.ID,
		Name:        latestLessPlan.Name,
		ContentType: v2.AssessmentContentTypeLessonPlan,
		OutcomeIDs:  latestLessPlan.OutcomeIDs,
		LatestID:    latestLessPlan.ID,
		FileType:    latestLessPlan.FileType,
	}
	asg.latestContentsFromSchedule = append(asg.latestContentsFromSchedule, lessPlan)

	// filling material
	dedupMap := make(map[string]struct{})
	subContents := subContentsMap[latestLessPlan.ID]
	for _, item := range subContents {
		if _, ok := dedupMap[item.ID]; ok {
			continue
		}
		subContentItem := &v2.AssessmentContentView{
			ID:          item.ID,
			Name:        item.Name,
			ContentType: v2.AssessmentContentTypeLessonMaterial,
			OutcomeIDs:  item.OutcomeIDs,
			LatestID:    item.ID,
			FileType:    item.FileType,
		}
		asg.latestContentsFromSchedule = append(asg.latestContentsFromSchedule, subContentItem)
		dedupMap[item.ID] = struct{}{}
	}

	// convert content outcome to latest outcome
	err = asg.ConvertContentOutcome(asg.latestContentsFromSchedule)
	if err != nil {
		return nil, err
	}
	asg.initRecord[SingleLatestContentSliceFromSchedule] = true

	return asg.latestContentsFromSchedule, nil
}

func (asg *AssessmentSingleGrainItem) SingleGetSchedule() (*entity.Schedule, error) {
	scheduleMap, err := asg.amg.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	schedule, ok := scheduleMap[asg.assessment.ScheduleID]
	if !ok {
		return nil, constant.ErrRecordNotFound
	}

	return schedule, nil
}

func (asg *AssessmentSingleGrainItem) SingleGetLockedContentsFromSchedule() ([]*v2.AssessmentContentView, error) {
	if asg.initRecord[SingleLockedContentSliceFromSchedule] {
		return asg.lockedContentsFromSchedule, nil
	}

	schedule, err := asg.SingleGetSchedule()
	if err != nil {
		return nil, err
	}

	asg.lockedContentsFromSchedule = make([]*v2.AssessmentContentView, 0)

	result, err := asg.getLockedContentBySchedule(schedule)
	if err != nil {
		return nil, err
	}

	asg.lockedContentsFromSchedule = result
	asg.initRecord[SingleLockedContentSliceFromSchedule] = true

	return asg.lockedContentsFromSchedule, nil
}

func (asg *AssessmentSingleGrainItem) getLockedContentBySchedule(schedule *entity.Schedule) ([]*v2.AssessmentContentView, error) {
	ctx := asg.ctx

	if !schedule.IsLockedLessonPlan() {
		log.Warn(ctx, "schedule not locked lesson plan", log.Any("schedule", schedule))
		return nil, constant.ErrInvalidArgs
	}

	dedupMap := make(map[string]struct{})

	// Extract and deduplicate contentID
	contentIDs := make([]string, 0)
	contentIDs = append(contentIDs, schedule.LiveLessonPlan.LessonPlanID)
	for _, materialItem := range schedule.LiveLessonPlan.LessonMaterials {
		if _, ok := dedupMap[materialItem.LessonMaterialID]; ok {
			continue
		}
		contentIDs = append(contentIDs, materialItem.LessonMaterialID)
		dedupMap[materialItem.LessonMaterialID] = struct{}{}
	}

	// get content info
	contents, err := GetContentModel().GetContentByIDListInternal(ctx, dbo.MustGetDB(ctx), contentIDs)
	if err != nil {
		log.Error(ctx, "toViews: GetContentModel().GetContentByIDList: get failed",
			log.Err(err),
			log.Strings("lesson_plan_ids", contentIDs),
		)
		return nil, err
	}
	contentMap := make(map[string]*entity.ContentInfoInternal, len(contents))
	for _, item := range contents {
		contentMap[item.ID] = item
	}

	// convert to content view
	result := make([]*v2.AssessmentContentView, 0, len(contents))
	if contentItem, ok := contentMap[schedule.LiveLessonPlan.LessonPlanID]; ok {
		resultItem := &v2.AssessmentContentView{
			ID:          contentItem.ID,
			Name:        contentItem.Name,
			ContentType: v2.AssessmentContentTypeLessonPlan,
			OutcomeIDs:  contentItem.OutcomeIDs,
			LatestID:    contentItem.LatestID,
			FileType:    contentItem.FileType,
		}

		result = append(result, resultItem)
	} else {
		log.Warn(ctx, "not found lessPlan", log.Any("contentMap", contentMap), log.Any("LiveLessonPlan", schedule.LiveLessonPlan))
		return nil, constant.ErrInvalidArgs
	}

	for _, materialItem := range schedule.LiveLessonPlan.LessonMaterials {
		contentItem, ok := contentMap[materialItem.LessonMaterialID]
		if !ok {
			log.Warn(ctx, "not found material", log.Any("contentMap", contentMap), log.String("LessonMaterialID", materialItem.LessonMaterialID))
			continue
		}
		resultItem := &v2.AssessmentContentView{
			ID:          contentItem.ID,
			Name:        contentItem.Name,
			ContentType: v2.AssessmentContentTypeLessonMaterial,
			OutcomeIDs:  contentItem.OutcomeIDs,
			LatestID:    contentItem.LatestID,
			FileType:    contentItem.FileType,
		}

		result = append(result, resultItem)
	}

	return result, nil
}

func (asg *AssessmentSingleGrainItem) SingleGetContentsFromSchedule() ([]*v2.AssessmentContentView, error) {
	var result []*v2.AssessmentContentView
	var err error
	if ok, _ := asg.IsNeedConvertLatestContent(); ok {
		result, err = asg.SingleGetLatestContentsFromSchedule()
	} else {
		result, err = asg.SingleGetLockedContentsFromSchedule()
	}

	if err != nil {
		return nil, err
	}

	if len(result) <= 0 {
		return nil, err
	}

	err = asg.ConvertContentOutcome(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (asg *AssessmentSingleGrainItem) SingleGetOutcomeMapFromContent() (map[string]*entity.Outcome, error) {
	if asg.initRecord[SingleOutcomeMapFromContent] {
		return asg.outcomeMapFromContent, nil
	}

	ctx := asg.ctx
	op := asg.op

	contents, err := asg.SingleGetContentsFromSchedule()
	if err != nil {
		return nil, err
	}

	outcomeIDs := make([]string, 0)
	deDupMap := make(map[string]struct{})

	for _, materialItem := range contents {
		for _, outcomeID := range materialItem.OutcomeIDs {
			if _, ok := deDupMap[outcomeID]; !ok {
				outcomeIDs = append(outcomeIDs, outcomeID)
				deDupMap[outcomeID] = struct{}{}
			}
		}
	}

	outcomes, err := GetOutcomeModel().GetByIDs(ctx, op, dbo.MustGetDB(ctx), outcomeIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*entity.Outcome, len(outcomes))

	for _, item := range outcomes {
		result[item.ID] = item
	}

	asg.initRecord[SingleOutcomeMapFromContent] = true
	asg.outcomeMapFromContent = result

	return result, nil
}

func (asg *AssessmentSingleGrainItem) SingleGetAssessmentContentMap() (map[string]*v2.AssessmentContent, error) {
	if asg.initRecord[SingleContentMapFromAssessment] {
		return asg.contentMapFromAssessment, nil
	}

	ctx := asg.ctx

	var assessmentContents []*v2.AssessmentContent
	err := assessmentV2.GetAssessmentContentDA().Query(ctx, &assessmentV2.AssessmentUserCondition{
		AssessmentID: sql.NullString{
			String: asg.assessment.ID,
			Valid:  true,
		},
	}, &assessmentContents)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*v2.AssessmentContent)
	if ok, _ := asg.IsNeedConvertLatestContent(); ok {
		oldContentIDs := make([]string, len(assessmentContents))
		assessmentContentMap := make(map[string]*v2.AssessmentContent)
		for i, item := range assessmentContents {
			oldContentIDs[i] = item.ContentID
			assessmentContentMap[item.ContentID] = item
		}

		latestContentIDMap, err := GetContentModel().GetLatestContentIDMapByIDListInternal(ctx, dbo.MustGetDB(ctx), oldContentIDs)
		if err != nil {
			log.Error(ctx, "GetLatestContentIDMapByIDList error", log.Err(err), log.Strings("oldContentIDs", oldContentIDs))
			return nil, err
		}

		for oldID, newID := range latestContentIDMap {
			result[newID] = assessmentContentMap[oldID]
		}
	} else {
		result = make(map[string]*v2.AssessmentContent, len(assessmentContents))
		for _, item := range assessmentContents {
			result[item.ContentID] = item
		}
	}

	asg.initRecord[SingleContentMapFromAssessment] = true
	asg.contentMapFromAssessment = result

	return result, nil
}

func (asg *AssessmentSingleGrainItem) SingleGetContentMapFromLiveRoom() (map[string]*RoomContentTree, error) {
	ctx := asg.ctx
	//op := adc.op

	if asg.initRecord[SingleContentMapFromLiveRoom] {
		return asg.contentMapFromLiveRoom, nil
	}

	_, roomContents, err := asg.SingleGetRoomData()
	if err != nil {
		return nil, err
	}

	result := make(map[string]*RoomContentTree, len(roomContents))
	if ok, _ := asg.IsNeedConvertLatestContent(); ok {
		oldContentIDs := make([]string, len(roomContents))
		for i, item := range roomContents {
			if item.TreeParentID == "" {
				oldContentIDs[i] = item.ContentID
			}
		}

		oldContentIDs = utils.SliceDeduplicationExcludeEmpty(oldContentIDs)
		latestContentIDMap, err := GetContentModel().GetLatestContentIDMapByIDListInternal(ctx, dbo.MustGetDB(ctx), oldContentIDs)
		if err != nil {
			log.Error(ctx, "GetLatestContentIDMapByIDList error", log.Err(err), log.Strings("oldContentIDs", oldContentIDs))
			return nil, err
		}

		for _, item := range roomContents {
			result[latestContentIDMap[item.ContentID]] = item
		}
	} else {
		for _, item := range roomContents {
			result[item.ContentID] = item
		}
	}

	asg.contentMapFromLiveRoom = result
	asg.initRecord[SingleContentMapFromLiveRoom] = true

	return result, nil
}

func (asg *AssessmentSingleGrainItem) SingleGetRoomData() (map[string][]*RoomUserScore, []*RoomContentTree, error) {
	ctx := asg.ctx
	//op := adc.op

	roomDataMap, err := asg.amg.GetRoomStudentScoresAndComments()
	if err != nil {
		return nil, nil, err
	}
	roomData, ok := roomDataMap[asg.assessment.ScheduleID]
	if !ok {
		log.Warn(ctx, "not found room data", log.Any("roomDataMap", roomDataMap), log.Any("assessment", asg.assessment))
		return make(map[string][]*RoomUserScore), nil, nil
	}

	return GetAssessmentExternalService().StudentScores(ctx, roomData.ScoresByUser)
}

func (asg *AssessmentSingleGrainItem) SingleGetCommentResultMap() (map[string]string, error) {
	if asg.initRecord[SingleUserCommentResultMap] {
		return asg.commentResultMap, nil
	}

	ctx := asg.ctx
	//op := asg.op

	asg.commentResultMap = make(map[string]string)

	studentRoomInfoMap, err := asg.amg.GetRoomStudentScoresAndComments()
	if err != nil {
		return nil, err
	}
	studentRoomInfo, ok := studentRoomInfoMap[asg.assessment.ScheduleID]
	if !ok {
		return asg.commentResultMap, nil
	}

	for _, item := range studentRoomInfo.TeacherCommentsByStudent {
		if item.User == nil {
			log.Warn(ctx, "get user comment error,user is empty", log.Any("studentRoomInfo", studentRoomInfo))
		}
		if len(item.TeacherComments) <= 0 {
			continue
		}
		latestComment := item.TeacherComments[len(item.TeacherComments)-1]

		asg.commentResultMap[item.User.UserID] = latestComment.Comment
	}

	asg.initRecord[SingleUserCommentResultMap] = true
	return asg.commentResultMap, nil
}

func (asg *AssessmentSingleGrainItem) SingleGetOutcomeFromAssessment() (map[string]*v2.AssessmentUserOutcome, error) {
	if asg.initRecord[SingleOutcomeMapFromAssessment] {
		return asg.outcomeMapFromAssessment, nil
	}
	ctx := asg.ctx

	assessmentUserMap, err := asg.amg.GetAssessmentUserMap()
	if err != nil {
		return nil, err
	}
	assessmentUsers, ok := assessmentUserMap[asg.assessment.ID]
	if !ok {
		return nil, constant.ErrRecordNotFound
	}

	assessmentUserPKs := make([]string, 0, len(assessmentUsers))
	for _, item := range assessmentUsers {
		if item.UserType == v2.AssessmentUserTypeStudent {
			assessmentUserPKs = append(assessmentUserPKs, item.ID)
		}
	}

	var userOutcomes []*v2.AssessmentUserOutcome
	userOutcomeCond := &assessmentV2.AssessmentUserOutcomeCondition{
		AssessmentUserIDs: entity.NullStrings{
			Strings: assessmentUserPKs,
			Valid:   true,
		},
	}
	err = assessmentV2.GetAssessmentUserOutcomeDA().Query(ctx, userOutcomeCond, &userOutcomes)
	if err != nil {
		log.Error(ctx, "query assessment user outcome error", log.Err(err), log.Any("userOutcomeCond", userOutcomeCond))
		return nil, err
	}

	result := make(map[string]*v2.AssessmentUserOutcome, len(userOutcomes))
	for _, item := range userOutcomes {
		key := asg.GetKey([]string{
			item.AssessmentUserID,
			item.AssessmentContentID,
			item.OutcomeID,
		})
		result[key] = item
	}

	asg.outcomeMapFromAssessment = result
	asg.initRecord[SingleOutcomeMapFromAssessment] = true

	return result, nil
}

func (asg *AssessmentSingleGrainItem) SingleGetContentsFromScheduleReview() (map[string]*entity.ScheduleReview, map[string]*entity.ContentInfoInternal, error) {
	ctx := asg.ctx
	op := asg.op

	studentReviews, err := GetScheduleModel().GetSuccessScheduleReview(ctx, op, asg.assessment.ScheduleID)
	if err != nil {
		return nil, nil, err
	}
	contentIDs := make([]string, 0)
	dedupContentID := make(map[string]struct{})
	result := make(map[string]*entity.ScheduleReview)
	for _, item := range studentReviews {
		if item.LiveLessonPlan == nil {
			continue
		}
		result[item.StudentID] = item
		for _, contentItem := range item.LiveLessonPlan.LessonMaterials {
			if _, ok := dedupContentID[contentItem.LessonMaterialID]; ok {
				continue
			}
			dedupContentID[contentItem.LessonMaterialID] = struct{}{}
			contentIDs = append(contentIDs, contentItem.LessonMaterialID)
		}
	}

	contents, err := GetContentModel().GetContentByIDListInternal(ctx, dbo.MustGetDB(ctx), contentIDs)
	if err != nil {
		return nil, nil, err
	}

	contentResult := make(map[string]*entity.ContentInfoInternal, len(contents))
	for _, item := range contents {
		contentResult[item.ID] = item
	}
	return result, contentResult, nil
}

func (asg *AssessmentSingleGrainItem) SingleGetOutcomesFromSchedule() ([]*entity.Outcome, error) {
	if asg.initRecord[SingleGetOutcomesFromSchedule] {
		return asg.outcomesFromSchedule, nil
	}

	ctx := asg.ctx
	op := asg.op

	outcomeIDMap, err := GetScheduleModel().GetLearningOutcomeIDs(ctx, op, []string{asg.assessment.ScheduleID})
	if err != nil {
		return nil, err
	}

	outcomeIDs, ok := outcomeIDMap[asg.assessment.ScheduleID]
	if !ok || len(outcomeIDs) <= 0 {
		return make([]*entity.Outcome, 0), nil
	}

	outcomes, err := GetOutcomeModel().GetByIDs(ctx, op, dbo.MustGetDB(ctx), outcomeIDs)
	if err != nil {
		return nil, err
	}

	asg.outcomesFromSchedule = outcomes
	asg.initRecord[SingleGetOutcomesFromSchedule] = true

	return outcomes, nil
}
