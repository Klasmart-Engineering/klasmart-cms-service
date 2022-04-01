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

	SingleOutcomeMapFromContent
	SingleLatestContentSliceFromSchedule
	SingleLockedContentSliceFromSchedule
	SingleContentMapFromAssessment
	SingleContentMapFromLiveRoom
	SingleUserCommentResultMap
	SingleOutcomeMapFromAssessment
)

type AssessmentGrain struct {
	ctx        context.Context
	op         *entity.Operator
	InitRecord map[AssessmentsGrainInit]bool

	assessments []*v2.Assessment
	assessment  *v2.Assessment

	AssessmentMulGrainItem

	AssessmentSingleGrainItem
}

type AssessmentMulGrainItem struct {
	assessmentMap        map[string]*v2.Assessment             // assessmentID
	scheduleMap          map[string]*entity.Schedule           // scheduleID
	scheduleRelationMap  map[string][]*entity.ScheduleRelation // scheduleID
	assessmentUserMap    map[string][]*v2.AssessmentUser       // assessmentID
	assessmentContentMap map[string][]*v2.AssessmentContent    // assessmentID
	programMap           map[string]*entity.IDName             // programID
	subjectMap           map[string]*entity.IDName             // subjectID
	classMap             map[string]*entity.IDName             // classID
	userMap              map[string]*entity.IDName             // userID
	liveRoomMap          map[string][]*external.H5PUserScores  // roomID
	lessPlanMap          map[string]*v2.AssessmentContentView  // lessPlanID

	assessmentUsers []*v2.AssessmentUser
}

type AssessmentSingleGrainItem struct {
	assessment *v2.Assessment

	outcomeMapFromContent      map[string]*entity.Outcome // key: outcomeID
	latestContentsFromSchedule []*v2.AssessmentContentView
	lockedContentsFromSchedule []*v2.AssessmentContentView
	contentMapFromAssessment   map[string]*v2.AssessmentContent     // key:contentID
	contentMapFromLiveRoom     map[string]*RoomContent              // key: contentID
	commentResultMap           map[string]string                    // userID
	outcomeMapFromAssessment   map[string]*v2.AssessmentUserOutcome // key: AssessmentUserID+AssessmentContentID+OutcomeID
}

func NewAssessmentGrainMul(ctx context.Context, op *entity.Operator, assessments []*v2.Assessment) *AssessmentGrain {
	return &AssessmentGrain{
		ctx:         ctx,
		op:          op,
		InitRecord:  make(map[AssessmentsGrainInit]bool),
		assessments: assessments,
		assessment:  new(v2.Assessment),
	}
}

func NewAssessmentGrainSingle(ctx context.Context, op *entity.Operator, assessment *v2.Assessment) *AssessmentGrain {
	return &AssessmentGrain{
		ctx:         ctx,
		op:          op,
		InitRecord:  make(map[AssessmentsGrainInit]bool),
		assessment:  assessment,
		assessments: []*v2.Assessment{assessment},
	}
}

func (ags *AssessmentGrain) GetAssessmentMap() (map[string]*v2.Assessment, error) {
	if ags.InitRecord[GrainAssessmentMap] {
		return ags.assessmentMap, nil
	}

	ags.assessmentMap = make(map[string]*v2.Assessment, len(ags.assessments))
	for _, item := range ags.assessments {
		ags.assessmentMap[item.ID] = item
	}

	ags.InitRecord[GrainAssessmentMap] = true

	return ags.assessmentMap, nil
}

func (ags *AssessmentGrain) GetScheduleMap() (map[string]*entity.Schedule, error) {
	if ags.InitRecord[GrainScheduleMap] {
		return ags.scheduleMap, nil
	}

	ctx := ags.ctx
	scheduleIDs := make([]string, len(ags.assessments))
	for i, item := range ags.assessments {
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

	ags.scheduleMap = make(map[string]*entity.Schedule, len(schedules))
	for _, item := range schedules {
		ags.scheduleMap[item.ID] = item
	}

	ags.InitRecord[GrainScheduleMap] = true

	return ags.scheduleMap, nil
}

func (ags *AssessmentGrain) GetScheduleRelationMap() (map[string][]*entity.ScheduleRelation, error) {
	if ags.InitRecord[GrainScheduleRelationMap] {
		return ags.scheduleRelationMap, nil
	}
	ctx := ags.ctx
	op := ags.op

	scheduleIDs := make([]string, len(ags.assessments))
	for i, item := range ags.assessments {
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

	ags.scheduleRelationMap = make(map[string][]*entity.ScheduleRelation, len(scheduleIDs))
	for _, item := range scheduleRelations {
		ags.scheduleRelationMap[item.ScheduleID] = append(ags.scheduleRelationMap[item.ScheduleID], item)
	}

	ags.InitRecord[GrainScheduleRelationMap] = true

	return ags.scheduleRelationMap, nil
}

func (ags *AssessmentGrain) GetAssessmentUsers() ([]*v2.AssessmentUser, error) {
	if ags.InitRecord[GrainAssessmentUserSlice] {
		return ags.assessmentUsers, nil
	}

	ctx := ags.ctx

	assessmentIDs := make([]string, len(ags.assessments))
	for i, item := range ags.assessments {
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

	ags.assessmentUsers = assessmentUsers
	ags.InitRecord[GrainAssessmentUserSlice] = true

	return ags.assessmentUsers, nil
}

func (ags *AssessmentGrain) GetAssessmentUserMap() (map[string][]*v2.AssessmentUser, error) {
	if ags.InitRecord[GrainAssessmentUserMap] {
		return ags.assessmentUserMap, nil
	}

	assessmentUsers, err := ags.GetAssessmentUsers()
	if err != nil {
		return nil, err
	}

	ags.assessmentUserMap = make(map[string][]*v2.AssessmentUser, len(ags.assessments))
	for _, item := range assessmentUsers {
		ags.assessmentUserMap[item.AssessmentID] = append(ags.assessmentUserMap[item.AssessmentID], item)
	}

	ags.InitRecord[GrainAssessmentUserMap] = true

	return ags.assessmentUserMap, nil
}

func (ags *AssessmentGrain) GetProgramMap() (map[string]*entity.IDName, error) {
	if ags.InitRecord[GrainProgramMap] {
		return ags.programMap, nil
	}

	ctx := ags.ctx
	op := ags.op

	scheduleMap, err := ags.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	programIDs := make([]string, 0)
	deDupMap := make(map[string]struct{})
	for _, item := range ags.assessments {
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

	ags.programMap = make(map[string]*entity.IDName)
	for _, item := range programs {
		if item == nil || item.ID == "" {
			log.Warn(ctx, "program id is empty", log.Any("programs", programs))
			continue
		}
		ags.programMap[item.ID] = &entity.IDName{
			ID:   item.ID,
			Name: item.Name,
		}
	}

	ags.InitRecord[GrainProgramMap] = true

	return ags.programMap, nil
}

func (ags *AssessmentGrain) GetSubjectMap() (map[string]*entity.IDName, error) {
	if ags.InitRecord[GrainSubjectMap] {
		return ags.subjectMap, nil
	}

	ctx := ags.ctx
	op := ags.op

	relationMap, err := ags.GetScheduleRelationMap()
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

	ags.subjectMap = make(map[string]*entity.IDName)
	for _, item := range subjects {
		if item == nil || item.ID == "" {
			log.Warn(ctx, "subject is empty", log.Any("subjects", subjects))
			continue
		}
		ags.subjectMap[item.ID] = &entity.IDName{
			ID:   item.ID,
			Name: item.Name,
		}
	}

	ags.InitRecord[GrainSubjectMap] = true

	return ags.subjectMap, nil
}

func (ags *AssessmentGrain) GetClassMap() (map[string]*entity.IDName, error) {
	if ags.InitRecord[GrainClassMap] {
		return ags.classMap, nil
	}

	ctx := ags.ctx
	op := ags.op

	relationMap, err := ags.GetScheduleRelationMap()
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

	ags.classMap = make(map[string]*entity.IDName)
	for _, item := range classes {
		if item == nil || item.ID == "" {
			continue
		}
		ags.classMap[item.ID] = &entity.IDName{
			ID:   item.ID,
			Name: item.Name,
		}
	}

	ags.InitRecord[GrainClassMap] = true

	return ags.classMap, nil
}

func (ags *AssessmentGrain) GetUserMap() (map[string]*entity.IDName, error) {
	if ags.InitRecord[GrainUserMap] {
		return ags.userMap, nil
	}

	ctx := ags.ctx
	op := ags.op

	assessmentUsers, err := ags.GetAssessmentUsers()
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

	ags.userMap = make(map[string]*entity.IDName)
	for _, item := range users {
		if !item.Valid {
			log.Warn(ctx, "user is inValid", log.Any("item", item))
			continue
		}
		ags.userMap[item.ID] = &entity.IDName{
			ID:   item.ID,
			Name: item.Name,
		}
	}

	ags.InitRecord[GrainUserMap] = true

	return ags.userMap, nil
}

func (ags *AssessmentGrain) GetRoomData() (map[string][]*external.H5PUserScores, error) {
	if ags.InitRecord[GrainLiveRoomMap] {
		return ags.liveRoomMap, nil
	}

	ctx := ags.ctx
	op := ags.op

	scheduleMap, err := ags.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	scheduleIDs := make([]string, 0, len(ags.assessments))
	for _, item := range scheduleMap {
		scheduleIDs = append(scheduleIDs, item.ID)
	}

	roomDataMap, err := external.GetH5PRoomScoreServiceProvider().BatchGet(ctx, op, scheduleIDs)
	if err != nil {
		log.Warn(ctx, "external service error",
			log.Err(err), log.Strings("scheduleIDs", scheduleIDs), log.Any("op", ags.op))
		ags.liveRoomMap = make(map[string][]*external.H5PUserScores)
	} else {
		ags.liveRoomMap = roomDataMap
	}

	ags.InitRecord[GrainLiveRoomMap] = true

	return ags.liveRoomMap, nil
}

func (ags *AssessmentGrain) GetLessPlanMap() (map[string]*v2.AssessmentContentView, error) {
	ctx := ags.ctx

	if ags.InitRecord[GrainLessPlanMap] {
		return ags.lessPlanMap, nil
	}

	scheduleMap, err := ags.GetScheduleMap()
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

	ags.lessPlanMap = make(map[string]*v2.AssessmentContentView, len(lessPlans))
	for _, item := range lessPlans {
		lessPlanItem := &v2.AssessmentContentView{
			ID:          item.ID,
			Name:        item.Name,
			OutcomeIDs:  item.OutcomeIDs,
			ContentType: v2.AssessmentContentTypeLessonPlan,
			LatestID:    item.LatestID,
			FileType:    item.FileType,
		}
		ags.lessPlanMap[item.ID] = lessPlanItem
	}

	// update schedule lessPlan ID
	for _, item := range scheduleMap {
		if item.IsLockedLessonPlan() {
			item.LessonPlanID = item.LiveLessonPlan.LessonPlanID
		} else {
			item.LessonPlanID = latestLassPlanIDMap[item.LessonPlanID]
		}
	}

	ags.InitRecord[GrainLessPlanMap] = true

	return ags.lessPlanMap, nil
}

func (ags *AssessmentGrain) GetAssessmentUserWithUserIDAndUserTypeMap() (map[string]*v2.AssessmentUser, error) {
	ctx := ags.ctx

	assessmentUserMap, err := ags.GetAssessmentUserMap()
	if err != nil {
		return nil, err
	}

	assessmentUsers, ok := assessmentUserMap[ags.assessment.ID]
	if !ok {
		log.Error(ctx, "not found assessment users", log.Any("assessment", ags.assessment), log.Any("assessmentUserMap", assessmentUserMap))
		return nil, constant.ErrRecordNotFound
	}

	result := make(map[string]*v2.AssessmentUser, len(assessmentUsers))
	for _, item := range assessmentUsers {
		result[ags.GetKey([]string{item.UserID, item.UserType.String()})] = item
	}

	return result, nil
}

func (asg *AssessmentGrain) GetKey(value []string) string {
	return strings.Join(value, "_")
}

func (asg *AssessmentGrain) IsNeedConvertLatestContent() (bool, error) {
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

func (asg *AssessmentGrain) ConvertContentOutcome(contents []*v2.AssessmentContentView) error {
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

func (asg *AssessmentGrain) SingleGetLatestContentsFromSchedule() ([]*v2.AssessmentContentView, error) {
	if asg.InitRecord[SingleLatestContentSliceFromSchedule] {
		return asg.latestContentsFromSchedule, nil
	}

	ctx := asg.ctx
	op := asg.op

	schedule, err := asg.SingleGetSchedule()
	if err != nil {
		return nil, err
	}

	latestLessPlanIDMap, err := GetContentModel().GetLatestContentIDMapByIDListInternal(ctx, dbo.MustGetDB(ctx), []string{schedule.LessonPlanID})
	if err != nil {
		return nil, err
	}
	latestLessPlanID, ok := latestLessPlanIDMap[schedule.LessonPlanID]
	if !ok {
		log.Error(ctx, "lessPlan not found", log.Any("schedule", schedule), log.Any("latestLessPlanIDMap", latestLessPlanIDMap))
		return nil, constant.ErrRecordNotFound
	}

	latestLessPlans, err := GetContentModel().GetContentByIDListInternal(ctx, dbo.MustGetDB(ctx), []string{latestLessPlanID})
	if err != nil {
		return nil, err
	}
	if len(latestLessPlans) <= 0 {
		return nil, constant.ErrRecordNotFound
	}

	subContentsMap, err := GetContentModel().GetContentsSubContentsMapByIDListInternal(ctx, dbo.MustGetDB(ctx), []string{latestLessPlanID}, op)
	if err != nil {
		return nil, err
	}

	latestLessPlan := latestLessPlans[0]
	subContents := subContentsMap[latestLessPlan.ID]

	lessPlan := &v2.AssessmentContentView{
		ID:          latestLessPlan.ID,
		Name:        latestLessPlan.Name,
		ContentType: v2.AssessmentContentTypeLessonPlan,
		OutcomeIDs:  latestLessPlan.OutcomeIDs,
		LatestID:    latestLessPlan.ID,
		FileType:    latestLessPlan.FileType,
	}
	asg.latestContentsFromSchedule = append(asg.latestContentsFromSchedule, lessPlan)

	for _, item := range subContents {
		subContentItem := &v2.AssessmentContentView{
			ID:          item.ID,
			Name:        item.Name,
			ContentType: v2.AssessmentContentTypeLessonMaterial,
			OutcomeIDs:  item.OutcomeIDs,
			LatestID:    item.ID,
			FileType:    item.FileType,
		}
		asg.latestContentsFromSchedule = append(asg.latestContentsFromSchedule, subContentItem)
	}

	asg.ConvertContentOutcome(asg.latestContentsFromSchedule)
	asg.InitRecord[SingleLatestContentSliceFromSchedule] = true

	return asg.latestContentsFromSchedule, nil
}

func (asg *AssessmentGrain) SingleGetSchedule() (*entity.Schedule, error) {
	scheduleMap, err := asg.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	schedule, ok := scheduleMap[asg.assessment.ScheduleID]
	if !ok {
		return nil, constant.ErrRecordNotFound
	}

	return schedule, nil
}

func (asg *AssessmentGrain) SingleGetLockedContentsFromSchedule() ([]*v2.AssessmentContentView, error) {
	if asg.InitRecord[SingleLockedContentSliceFromSchedule] {
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
	asg.InitRecord[SingleLockedContentSliceFromSchedule] = true

	return asg.lockedContentsFromSchedule, nil
}

func (asg *AssessmentGrain) getLockedContentBySchedule(schedule *entity.Schedule) ([]*v2.AssessmentContentView, error) {
	ctx := asg.ctx
	contentIDs := make([]string, 0)
	contentIDs = append(contentIDs, schedule.LiveLessonPlan.LessonPlanID)
	for _, materialItem := range schedule.LiveLessonPlan.LessonMaterials {
		contentIDs = append(contentIDs, materialItem.LessonMaterialID)
	}

	contentIDs = utils.SliceDeduplication(contentIDs)
	contents, err := GetContentModel().GetContentByIDListInternal(ctx, dbo.MustGetDB(ctx), contentIDs)
	if err != nil {
		log.Error(ctx, "toViews: GetContentModel().GetContentByIDList: get failed",
			log.Err(err),
			log.Strings("lesson_plan_ids", contentIDs),
		)
		return nil, err
	}
	contentOutcomeIDsMap := make(map[string][]string, len(contents))
	for _, item := range contents {
		contentOutcomeIDsMap[item.ID] = item.OutcomeIDs
	}

	contentInfoMap := make(map[string]*entity.ContentInfoInternal, len(contents))
	for _, item := range contents {
		contentInfoMap[item.ID] = item
	}

	liveLessonPlan := schedule.LiveLessonPlan

	lessPlan := &v2.AssessmentContentView{
		ID:          liveLessonPlan.LessonPlanID,
		Name:        liveLessonPlan.LessonPlanName,
		ContentType: v2.AssessmentContentTypeLessonPlan,
		OutcomeIDs:  contentOutcomeIDsMap[liveLessonPlan.LessonPlanID],
	}
	if contentItem, ok := contentInfoMap[liveLessonPlan.LessonPlanID]; ok {
		lessPlan.LatestID = contentItem.LatestID
		lessPlan.FileType = contentItem.FileType
	}

	result := append(asg.lockedContentsFromSchedule, lessPlan)

	for _, item := range liveLessonPlan.LessonMaterials {
		materialItem := &v2.AssessmentContentView{
			ID:          item.LessonMaterialID,
			Name:        item.LessonMaterialName,
			ContentType: v2.AssessmentContentTypeLessonMaterial,
			OutcomeIDs:  contentOutcomeIDsMap[item.LessonMaterialID],
		}
		if contentItem, ok := contentInfoMap[item.LessonMaterialID]; ok {
			materialItem.LatestID = contentItem.LatestID
			materialItem.FileType = contentItem.FileType
		}
		result = append(result, materialItem)
	}

	return result, nil
}

func (asg *AssessmentGrain) SingleGetContentsFromSchedule() ([]*v2.AssessmentContentView, error) {
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

func (asg *AssessmentGrain) SingleGetOutcomeMapFromContent() (map[string]*entity.Outcome, error) {
	if asg.InitRecord[SingleOutcomeMapFromContent] {
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

	asg.InitRecord[SingleOutcomeMapFromContent] = true
	asg.outcomeMapFromContent = result

	return result, nil
}

func (asg *AssessmentGrain) SingleGetAssessmentContentMap() (map[string]*v2.AssessmentContent, error) {
	if asg.InitRecord[SingleContentMapFromAssessment] {
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

	asg.InitRecord[SingleContentMapFromAssessment] = true
	asg.contentMapFromAssessment = result

	return result, nil
}

func (asg *AssessmentGrain) SingleGetContentMapFromLiveRoom() (map[string]*RoomContent, error) {
	ctx := asg.ctx
	//op := adc.op

	if asg.InitRecord[SingleContentMapFromLiveRoom] {
		return asg.contentMapFromLiveRoom, nil
	}

	roomInfo, err := asg.SingleGetRoomData()
	if err != nil {
		return nil, err
	}

	result := make(map[string]*RoomContent, len(roomInfo.Contents))
	if ok, _ := asg.IsNeedConvertLatestContent(); ok {
		oldContentIDs := make([]string, len(roomInfo.Contents))
		oldContentMap := make(map[string]*RoomContent)
		for i, item := range roomInfo.Contents {
			oldContentIDs[i] = item.MaterialID
			oldContentMap[item.MaterialID] = item
		}

		latestContentIDMap, err := GetContentModel().GetLatestContentIDMapByIDListInternal(ctx, dbo.MustGetDB(ctx), oldContentIDs)
		if err != nil {
			log.Error(ctx, "GetLatestContentIDMapByIDList error", log.Err(err), log.Strings("oldContentIDs", oldContentIDs))
			return nil, err
		}

		for _, item := range roomInfo.Contents {
			result[latestContentIDMap[item.MaterialID]] = item
		}
	} else {
		for _, item := range roomInfo.Contents {
			result[item.MaterialID] = item
		}
	}

	asg.contentMapFromLiveRoom = result
	asg.InitRecord[SingleContentMapFromLiveRoom] = true

	return result, nil
}

func (asg *AssessmentGrain) SingleGetRoomData() (*RoomInfo, error) {
	ctx := asg.ctx
	//op := adc.op

	roomDataMap, err := asg.GetRoomData()
	if err != nil {
		return nil, err
	}
	roomData, ok := roomDataMap[asg.assessment.ScheduleID]
	if !ok {
		log.Warn(ctx, "not found room data", log.Any("roomDataMap", roomDataMap), log.Any("assessment", asg.assessment))
		return new(RoomInfo), nil
	}

	roomResultInfo, err := getAssessmentLiveRoom().getRoomResultInfo(ctx, roomData)
	if err != nil {
		return nil, err
	}

	return roomResultInfo, nil
}

func (asg *AssessmentGrain) SingleGetCommentResultMap() (map[string]string, error) {
	if asg.InitRecord[SingleUserCommentResultMap] {
		return asg.commentResultMap, nil
	}

	ctx := asg.ctx
	op := asg.op

	asg.commentResultMap = make(map[string]string)

	commentResults, err := getAssessmentLiveRoom().batchGetRoomCommentMap(ctx, op, []string{asg.assessment.ScheduleID})
	if err != nil {
		log.Error(asg.ctx, "get assessment comment from live room error", log.Err(err), log.String("scheduleID", asg.assessment.ScheduleID))
	} else {
		if commentItem, ok := commentResults[asg.assessment.ScheduleID]; ok && commentItem != nil {
			asg.commentResultMap = commentItem
		}
	}

	asg.InitRecord[SingleUserCommentResultMap] = true
	return asg.commentResultMap, nil
}

func (asg *AssessmentGrain) SingleGetOutcomeFromAssessment() (map[string]*v2.AssessmentUserOutcome, error) {
	if asg.InitRecord[SingleOutcomeMapFromAssessment] {
		return asg.outcomeMapFromAssessment, nil
	}
	ctx := asg.ctx

	assessmentUserMap, err := asg.GetAssessmentUserMap()
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
	asg.InitRecord[SingleOutcomeMapFromAssessment] = true

	return result, nil
}

func (asg *AssessmentGrain) SingleGetContentsFromScheduleReview() (map[string]*entity.ScheduleReview, map[string]*entity.ContentInfoInternal, error) {
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
