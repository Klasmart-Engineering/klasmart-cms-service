package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da/assessmentV2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type AssessmentsGrainInit int

const (
	_ AssessmentsGrainInit = iota
	GrainAssessment
	GrainScheduleRelation
	GrainAssessmentUser
	GrainAssessmentContent
	GrainAssessmentUserSlice

	GrainSchedule
	GrainProgram
	GrainSubject
	GrainClass
	GrainUser
	GrainLiveRoom
	GrainLessPlan
)

type AssessmentsGrain struct {
	ctx        context.Context
	op         *entity.Operator
	InitRecord map[AssessmentsGrainInit]bool

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
	liveRoomMap          map[string][]*external.H5PUserScores  // roomID
	lessPlanMap          map[string]*v2.AssessmentContentView  // lessPlanID

	assessmentUsers []*v2.AssessmentUser
}

func NewAssessmentsGrain(ctx context.Context, op *entity.Operator, assessments []*v2.Assessment) *AssessmentsGrain {
	return &AssessmentsGrain{
		ctx:         ctx,
		op:          op,
		InitRecord:  make(map[AssessmentsGrainInit]bool),
		assessments: assessments,
	}
}

func (ags *AssessmentsGrain) GetAssessmentMap() (map[string]*v2.Assessment, error) {
	if ags.InitRecord[GrainAssessment] {
		return ags.assessmentMap, nil
	}

	ags.assessmentMap = make(map[string]*v2.Assessment, len(ags.assessments))
	for _, item := range ags.assessments {
		ags.assessmentMap[item.ID] = item
	}

	ags.InitRecord[GrainAssessment] = true

	return ags.assessmentMap, nil
}

func (ags *AssessmentsGrain) GetScheduleMap() (map[string]*entity.Schedule, error) {
	if ags.InitRecord[GrainSchedule] {
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

	ags.InitRecord[GrainSchedule] = true

	return ags.scheduleMap, nil
}

func (ags *AssessmentsGrain) GetScheduleRelationMap() (map[string][]*entity.ScheduleRelation, error) {
	if ags.InitRecord[GrainScheduleRelation] {
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

	ags.InitRecord[GrainScheduleRelation] = true

	return ags.scheduleRelationMap, nil
}

func (ags *AssessmentsGrain) GetAssessmentUsers() ([]*v2.AssessmentUser, error) {
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

func (ags *AssessmentsGrain) GetAssessmentUserMap() (map[string][]*v2.AssessmentUser, error) {
	if ags.InitRecord[GrainAssessmentUser] {
		return ags.assessmentUserMap, nil
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

	ags.assessmentUserMap = make(map[string][]*v2.AssessmentUser, len(ags.assessments))
	for _, item := range assessmentUsers {
		ags.assessmentUserMap[item.AssessmentID] = append(ags.assessmentUserMap[item.AssessmentID], item)
	}

	ags.InitRecord[GrainAssessmentUser] = true

	return ags.assessmentUserMap, nil
}

func (ags *AssessmentsGrain) GetAssessmentContentMap() (map[string][]*v2.AssessmentContent, error) {
	if ags.InitRecord[GrainAssessmentContent] {
		return ags.assessmentContentMap, nil
	}

	ctx := ags.ctx

	assessmentIDs := make([]string, len(ags.assessments))
	for i, item := range ags.assessments {
		assessmentIDs[i] = item.ID
	}

	var assessmentContents []*v2.AssessmentContent
	err := assessmentV2.GetAssessmentContentDA().Query(ctx, &assessmentV2.AssessmentUserCondition{
		AssessmentIDs: entity.NullStrings{
			Strings: assessmentIDs,
			Valid:   true,
		},
	}, &assessmentContents)
	if err != nil {
		return nil, err
	}

	for _, item := range assessmentContents {
		ags.assessmentContentMap[item.AssessmentID] = append(ags.assessmentContentMap[item.AssessmentID], item)
	}

	ags.InitRecord[GrainAssessmentContent] = true

	return ags.assessmentContentMap, nil
}

func (ags *AssessmentsGrain) GetProgramMap() (map[string]*entity.IDName, error) {
	if ags.InitRecord[GrainProgram] {
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

	ags.InitRecord[GrainProgram] = true

	return ags.programMap, nil
}

func (ags *AssessmentsGrain) GetSubjectMap() (map[string]*entity.IDName, error) {
	if ags.InitRecord[GrainSubject] {
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

	ags.InitRecord[GrainSubject] = true

	return ags.subjectMap, nil
}

func (ags *AssessmentsGrain) GetClassMap() (map[string]*entity.IDName, error) {
	if ags.InitRecord[GrainClass] {
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

	ags.InitRecord[GrainClass] = true

	return ags.classMap, nil
}

func (ags *AssessmentsGrain) GetUserMap(assessmentUserMap map[string][]*v2.AssessmentUser) (map[string]*entity.IDName, error) {
	if ags.InitRecord[GrainUser] {
		return ags.userMap, nil
	}

	ctx := ags.ctx
	op := ags.op

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

	ags.InitRecord[GrainUser] = true

	return ags.userMap, nil
}

func (ags *AssessmentsGrain) GetRoomData() (map[string][]*external.H5PUserScores, error) {
	if ags.InitRecord[GrainLiveRoom] {
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
		if item.LessonPlanID != "" {
			scheduleIDs = append(scheduleIDs, item.ID)
		}
	}

	roomDataMap, err := external.GetH5PRoomScoreServiceProvider().BatchGet(ctx, op, scheduleIDs)
	if err != nil {
		log.Warn(ctx, "external service error",
			log.Err(err), log.Strings("scheduleIDs", scheduleIDs), log.Any("op", ags.op))
		ags.liveRoomMap = make(map[string][]*external.H5PUserScores)
	} else {
		ags.liveRoomMap = roomDataMap
	}

	ags.InitRecord[GrainLiveRoom] = true

	return ags.liveRoomMap, nil
}

func (ags *AssessmentsGrain) GetLessPlanMap() (map[string]*v2.AssessmentContentView, error) {
	ctx := ags.ctx

	if ags.InitRecord[GrainLessPlan] {
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

	ags.InitRecord[GrainLessPlan] = true

	return ags.lessPlanMap, nil
}
