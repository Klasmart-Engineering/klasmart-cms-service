package model

import (
	"context"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da/assessmentV2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type AssessmentPageComponent struct {
	ctx context.Context
	op  *entity.Operator

	assessments []*v2.Assessment

	assScheduleMap      map[string]*entity.Schedule          // key:assessmentID
	assProgramMap       map[string]*entity.IDName            // key:assessmentID
	assLessPlanMap      map[string]*v2.AssessmentContentView // key:assessmentID
	assClassMap         map[string]*entity.IDName            // key:assessmentID
	assSubjectMap       map[string][]*entity.IDName          // key:assessmentID
	assTeacherMap       map[string][]*entity.IDName          // key:assessmentID
	assRemainingTimeMap map[string]int64                     // key:assessmentID
	assCompleteRateMap  map[string]float64                   // key:assessmentID

	allScheduleMap         map[string]*entity.Schedule           // key:scheduleID
	allProgramMap          map[string]*entity.IDName             // key:programID
	allSubjectMap          map[string]*entity.IDName             // key:subjectID
	allUserMap             map[string]*entity.IDName             // key:userID
	allLessPlanMap         map[string]*v2.AssessmentContentView  // key:lessPlanID
	allLessPlanWithMap     map[string]*v2.AssessmentContentView  // key:lessPlanID
	allClassMap            map[string]*entity.IDName             // key:classID
	allScheduleRelationMap map[string][]*entity.ScheduleRelation // key:schedule
	allRoomDataMap         map[string][]*external.H5PUserScores  // key:scheduleID

	assessmentUserMap    map[string][]*v2.AssessmentUser    // key:assessmentID
	assessmentContentMap map[string][]*v2.AssessmentContent // key:assessmentID
}

func NewPageComponent(ctx context.Context, op *entity.Operator, assessments []*v2.Assessment) *AssessmentPageComponent {
	return &AssessmentPageComponent{
		ctx:         ctx,
		op:          op,
		assessments: assessments,

		assScheduleMap:      make(map[string]*entity.Schedule),
		assProgramMap:       make(map[string]*entity.IDName),
		assLessPlanMap:      make(map[string]*v2.AssessmentContentView),
		assClassMap:         make(map[string]*entity.IDName),
		assSubjectMap:       make(map[string][]*entity.IDName),
		assTeacherMap:       make(map[string][]*entity.IDName),
		assRemainingTimeMap: make(map[string]int64),
		assCompleteRateMap:  make(map[string]float64),

		allScheduleMap:         make(map[string]*entity.Schedule),
		allProgramMap:          make(map[string]*entity.IDName),
		allSubjectMap:          make(map[string]*entity.IDName),
		allUserMap:             make(map[string]*entity.IDName),
		allLessPlanMap:         make(map[string]*v2.AssessmentContentView),
		allLessPlanWithMap:     make(map[string]*v2.AssessmentContentView),
		allClassMap:            make(map[string]*entity.IDName),
		allScheduleRelationMap: make(map[string][]*entity.ScheduleRelation),
		allRoomDataMap:         make(map[string][]*external.H5PUserScores),

		assessmentUserMap:    make(map[string][]*v2.AssessmentUser),
		assessmentContentMap: make(map[string][]*v2.AssessmentContent),
	}
}

func (apc *AssessmentPageComponent) GetScheduleMap() (map[string]*entity.Schedule, error) {
	ctx := apc.ctx

	if _, ok := apc.allScheduleMap[constant.AssessmentInitializedKey]; ok {
		return apc.allScheduleMap, nil
	}

	scheduleIDs := make([]string, len(apc.assessments))
	for i, item := range apc.assessments {
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

	apc.allScheduleMap = make(map[string]*entity.Schedule, len(schedules))
	for _, item := range schedules {
		apc.allScheduleMap[item.ID] = item
	}

	apc.allScheduleMap[constant.AssessmentInitializedKey] = &entity.Schedule{}

	return apc.allScheduleMap, nil
}

func (apc *AssessmentPageComponent) GetScheduleRelationMap() (map[string][]*entity.ScheduleRelation, error) {
	ctx := apc.ctx

	if len(apc.allScheduleRelationMap) > 0 {
		return apc.allScheduleRelationMap, nil
	}

	scheduleIDs := make([]string, len(apc.assessments))
	for i, item := range apc.assessments {
		scheduleIDs[i] = item.ScheduleID
	}

	scheduleRelations, err := GetScheduleRelationModel().Query(ctx, apc.op, &da.ScheduleRelationCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
	})
	if err != nil {
		return nil, err
	}

	apc.allScheduleRelationMap = make(map[string][]*entity.ScheduleRelation, len(scheduleIDs))
	for _, item := range scheduleRelations {
		apc.allScheduleRelationMap[item.ScheduleID] = append(apc.allScheduleRelationMap[item.ScheduleID], item)
	}

	if len(apc.allScheduleRelationMap) <= 0 {
		apc.allScheduleRelationMap[constant.AssessmentInitializedKey] = nil
	}

	return apc.allScheduleRelationMap, nil
}

//func (apc *AssessmentPageComponent) GetLessPlanMap() (map[string]*v2.AssessmentContentView, error) {
//	data, err := apc.prepareLessPlanMap()
//	if err != nil {
//		return nil, err
//	}
//
//	if len(data) == 1 {
//		return nil, constant.ErrRecordNotFound
//	}
//
//	return data, nil
//}

func (apc *AssessmentPageComponent) GetAssessmentUserMap() (map[string][]*v2.AssessmentUser, error) {
	if _, ok := apc.assessmentUserMap[constant.AssessmentInitializedKey]; ok {
		return apc.assessmentUserMap, nil
	}

	ctx := apc.ctx

	assessmentIDs := make([]string, len(apc.assessments))
	for i, item := range apc.assessments {
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

	for _, item := range assessmentUsers {
		apc.assessmentUserMap[item.AssessmentID] = append(apc.assessmentUserMap[item.AssessmentID], item)
	}

	apc.assessmentUserMap[constant.AssessmentInitializedKey] = nil

	return apc.assessmentUserMap, nil
}

func (apc *AssessmentPageComponent) GetAssessmentContentMap() (map[string][]*v2.AssessmentContent, error) {
	if _, ok := apc.assessmentContentMap[constant.AssessmentInitializedKey]; ok {
		return apc.assessmentContentMap, nil
	}

	ctx := apc.ctx

	assessmentIDs := make([]string, len(apc.assessments))
	for i, item := range apc.assessments {
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
		apc.assessmentContentMap[item.AssessmentID] = append(apc.assessmentContentMap[item.AssessmentID], item)
	}

	apc.assessmentContentMap[constant.AssessmentInitializedKey] = nil

	return apc.assessmentContentMap, nil
}

func (apc *AssessmentPageComponent) GetLatestContentIDMapByIDList(cids []string) (map[string]string, error) {
	ctx := apc.ctx

	resp := make(map[string]string)
	data, err := da.GetContentDA().GetContentByIDList(ctx, dbo.MustGetDB(ctx), cids)
	if err != nil {
		log.Error(ctx, "can't search content", log.Err(err), log.Strings("cids", cids))
		return nil, err
	}
	for i := range data {
		latestID := data[i].LatestID
		if data[i].LatestID == "" {
			latestID = data[i].ID
		}
		resp[data[i].ID] = latestID
	}
	return resp, nil
}

func (apc *AssessmentPageComponent) GetLessPlanMap() (map[string]*v2.AssessmentContentView, error) {
	ctx := apc.ctx
	tx := dbo.MustGetDB(ctx)

	if _, ok := apc.allLessPlanMap[constant.AssessmentInitializedKey]; ok {
		return apc.allLessPlanMap, nil
	}

	scheduleMap, err := apc.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	attemptedLessPlanIDs := make([]string, 0)
	notAttemptedLessPlanIDs := make([]string, 0)
	for _, item := range scheduleMap {
		if item.IsLockedLessonPlan() {
			attemptedLessPlanIDs = append(attemptedLessPlanIDs, item.LiveLessonPlan.LessonPlanID)
		} else {
			notAttemptedLessPlanIDs = append(notAttemptedLessPlanIDs, item.LessonPlanID)
		}
	}

	latestLassPlanIDMap, err := apc.GetLatestContentIDMapByIDList(notAttemptedLessPlanIDs)
	if err != nil {
		return nil, err
	}

	for _, latestID := range latestLassPlanIDMap {
		attemptedLessPlanIDs = append(attemptedLessPlanIDs, latestID)
	}

	lessPlanIDs := utils.SliceDeduplicationExcludeEmpty(attemptedLessPlanIDs)

	lessPlans, err := GetContentModel().GetContentNameByIDList(ctx, tx, lessPlanIDs)
	if err != nil {
		log.Error(ctx, "get content by ids error", log.Err(err), log.Strings("lessPlanIDs", lessPlanIDs))
		return nil, err
	}

	apc.allLessPlanMap = make(map[string]*v2.AssessmentContentView, len(lessPlans))
	for _, item := range lessPlans {
		lessPlanItem := &v2.AssessmentContentView{
			ID:          item.ID,
			Name:        item.Name,
			OutcomeIDs:  item.OutcomeIDs,
			ContentType: v2.AssessmentContentTypeLessonPlan,
			LatestID:    item.LatestID,
		}
		apc.allLessPlanMap[item.ID] = lessPlanItem
	}

	// update schedule lessPlan ID
	for _, item := range scheduleMap {
		if item.IsLockedLessonPlan() {
			item.LessonPlanID = item.LiveLessonPlan.LessonPlanID
		} else {
			item.LessonPlanID = latestLassPlanIDMap[item.LessonPlanID]
		}
	}

	apc.allLessPlanMap[constant.AssessmentInitializedKey] = new(v2.AssessmentContentView)

	return apc.allLessPlanMap, nil
}

func (apc *AssessmentPageComponent) GetProgramMap() (map[string]*entity.IDName, error) {
	ctx := apc.ctx
	op := apc.op

	if len(apc.allProgramMap) > 0 {
		return apc.allProgramMap, nil
	}

	scheduleMap, err := apc.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	programIDs := make([]string, 0)
	deDupMap := make(map[string]struct{})

	for _, item := range apc.assessments {
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
	for _, item := range programs {
		if item == nil || item.ID == "" {
			log.Warn(ctx, "program id is empty", log.Any("programs", programs))
			continue
		}
		apc.allProgramMap[item.ID] = &entity.IDName{
			ID:   item.ID,
			Name: item.Name,
		}
	}

	if len(apc.allProgramMap) <= 0 {
		apc.allProgramMap[constant.AssessmentInitializedKey] = new(entity.IDName)
	}

	return apc.allProgramMap, nil
}

func (apc *AssessmentPageComponent) GetSubjectMap() (map[string]*entity.IDName, error) {
	if len(apc.allSubjectMap) > 0 {
		return apc.allSubjectMap, nil
	}

	ctx := apc.ctx
	op := apc.op

	relationMap, err := apc.GetScheduleRelationMap()
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

	for _, item := range subjects {
		if item == nil || item.ID == "" {
			log.Warn(ctx, "subject is empty", log.Any("subjects", subjects))
			continue
		}
		apc.allSubjectMap[item.ID] = &entity.IDName{
			ID:   item.ID,
			Name: item.Name,
		}
	}

	if len(apc.allSubjectMap) <= 0 {
		apc.allSubjectMap[constant.AssessmentInitializedKey] = new(entity.IDName)
	}

	return apc.allSubjectMap, nil
}

func (apc *AssessmentPageComponent) GetClassMap() (map[string]*entity.IDName, error) {
	if len(apc.allClassMap) > 0 {
		return apc.allClassMap, nil
	}

	ctx := apc.ctx
	op := apc.op

	relationMap, err := apc.GetScheduleRelationMap()
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

	for _, item := range classes {
		if item == nil || item.ID == "" {
			continue
		}
		apc.allClassMap[item.ID] = &entity.IDName{
			ID:   item.ID,
			Name: item.Name,
		}
	}

	if len(apc.allClassMap) <= 0 {
		apc.allClassMap[constant.AssessmentInitializedKey] = new(entity.IDName)
	}

	return apc.allClassMap, nil
}

func (apc *AssessmentPageComponent) GetUserMap() (map[string]*entity.IDName, error) {
	if _, ok := apc.allUserMap[constant.AssessmentInitializedKey]; ok {
		return apc.allUserMap, nil
	}

	ctx := apc.ctx
	op := apc.op

	assessmentUserMap, err := apc.GetAssessmentUserMap()
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(assessmentUserMap))
	deDupMap := make(map[string]struct{})

	for _, auItems := range assessmentUserMap {
		for _, auItem := range auItems {
			if _, ok := deDupMap[auItem.UserID]; !ok {
				deDupMap[auItem.UserID] = struct{}{}
				userIDs = append(userIDs, auItem.UserID)
			}
		}
	}

	users, err := external.GetUserServiceProvider().BatchGet(ctx, op, userIDs)
	if err != nil {
		return nil, err
	}

	for _, item := range users {
		if !item.Valid {
			log.Warn(ctx, "user is inValid", log.Any("item", item))
			continue
		}
		apc.allUserMap[item.ID] = &entity.IDName{
			ID:   item.ID,
			Name: item.Name,
		}
	}

	apc.allUserMap[constant.AssessmentInitializedKey] = new(entity.IDName)

	return apc.allUserMap, nil
}

func (apc *AssessmentPageComponent) GetRoomData() (map[string][]*external.H5PUserScores, error) {
	if _, ok := apc.allRoomDataMap[constant.AssessmentInitializedKey]; ok {
		return apc.allRoomDataMap, nil
	}

	scheduleIDs := make([]string, 0, len(apc.assessments))
	for _, item := range apc.allScheduleMap {
		if item.LessonPlanID != "" {
			scheduleIDs = append(scheduleIDs, item.ID)
		}
	}

	roomDataMap, err := external.GetH5PRoomScoreServiceProvider().BatchGet(apc.ctx, apc.op, scheduleIDs)
	if err != nil {
		log.Warn(apc.ctx, "external service error", log.Err(err), log.Strings("scheduleIDs", scheduleIDs), log.Any("op", apc.op))
	} else {
		apc.allRoomDataMap = roomDataMap
	}

	apc.allRoomDataMap[constant.AssessmentInitializedKey] = nil

	return apc.allRoomDataMap, nil
}

// key: assessmentID
func (apc *AssessmentPageComponent) MatchSchedule() error {
	scheduleMap, err := apc.GetScheduleMap()
	if err != nil {
		return err
	}

	for _, item := range apc.assessments {
		if scheduleItem, ok := scheduleMap[item.ScheduleID]; ok {
			apc.assScheduleMap[item.ID] = scheduleItem
		}
	}

	return nil
}

// key: assessmentID
func (apc *AssessmentPageComponent) MatchLessPlan() error {
	if len(apc.assLessPlanMap) > 0 {
		return nil
	}

	scheduleMap, err := apc.GetScheduleMap()
	if err != nil {
		return err
	}

	lessPlanMap, err := apc.GetLessPlanMap()
	if err != nil {
		return err
	}

	apc.assLessPlanMap = make(map[string]*v2.AssessmentContentView, len(apc.assessments))
	for _, item := range apc.assessments {
		if schedule, ok := scheduleMap[item.ScheduleID]; ok {
			apc.assLessPlanMap[item.ID] = lessPlanMap[schedule.LessonPlanID]
		}
	}

	return nil
}

// key: assessmentID
func (apc *AssessmentPageComponent) MatchProgram() error {
	if len(apc.assProgramMap) > 0 {
		return nil
	}

	scheduleMap, err := apc.GetScheduleMap()
	if err != nil {
		return err
	}

	programMap, err := apc.GetProgramMap()
	if err != nil {
		return err
	}

	apc.assProgramMap = make(map[string]*entity.IDName, len(apc.assessments))
	for _, item := range apc.assessments {
		if schedule, ok := scheduleMap[item.ScheduleID]; ok {
			apc.assProgramMap[item.ID] = programMap[schedule.ProgramID]
		}
	}

	return nil
}

// key: assessmentID
func (apc *AssessmentPageComponent) MatchSubject() error {
	if len(apc.assSubjectMap) > 0 {
		return nil
	}

	relationMap, err := apc.GetScheduleRelationMap()
	if err != nil {
		return err
	}

	subjectMap, err := apc.GetSubjectMap()
	if err != nil {
		return err
	}

	apc.assSubjectMap = make(map[string][]*entity.IDName, len(apc.assessments))
	for _, item := range apc.assessments {
		if srItems, ok := relationMap[item.ScheduleID]; ok {
			for _, srItem := range srItems {
				if srItem.RelationType != entity.ScheduleRelationTypeSubject {
					continue
				}
				if subItem, ok := subjectMap[srItem.RelationID]; ok && subItem != nil {
					apc.assSubjectMap[item.ID] = append(apc.assSubjectMap[item.ID], subItem)
				}
			}
		}
	}

	return nil
}

// key: assessmentID
func (apc *AssessmentPageComponent) MatchTeacher() error {
	if len(apc.assTeacherMap) > 0 {
		return nil
	}

	assessmentUserMap, err := apc.GetAssessmentUserMap()
	if err != nil {
		return err
	}

	userMap, err := apc.GetUserMap()
	if err != nil {
		return err
	}

	apc.assTeacherMap = make(map[string][]*entity.IDName, len(apc.assessments))
	for _, item := range apc.assessments {
		if assUserItems, ok := assessmentUserMap[item.ID]; ok {
			for _, assUserItem := range assUserItems {
				if assUserItem.UserType != v2.AssessmentUserTypeTeacher {
					continue
				}
				if item.AssessmentType == v2.AssessmentTypeOnlineClass && assUserItem.StatusByUser == v2.AssessmentUserStatusNotParticipate {
					continue
				}

				if userItem, ok := userMap[assUserItem.UserID]; ok && userItem != nil {
					apc.assTeacherMap[item.ID] = append(apc.assTeacherMap[item.ID], userItem)
				}
			}
		}
	}

	return nil
}

// key: assessmentID
func (apc *AssessmentPageComponent) MatchClass() error {
	if len(apc.assClassMap) > 0 {
		return nil
	}

	relationMap, err := apc.GetScheduleRelationMap()
	if err != nil {
		return err
	}

	classMap, err := apc.GetClassMap()
	if err != nil {
		return err
	}

	apc.assClassMap = make(map[string]*entity.IDName, len(apc.assessments))
	for _, item := range apc.assessments {
		if srItems, ok := relationMap[item.ScheduleID]; ok {
			for _, srItem := range srItems {
				if srItem.RelationType == entity.ScheduleRelationTypeClassRosterClass {
					apc.assClassMap[item.ID] = classMap[srItem.RelationID]
					break
				}
			}
		}
	}

	return nil
}

func (apc *AssessmentPageComponent) MatchRemainingTime() error {
	if len(apc.assRemainingTimeMap) > 0 {
		return nil
	}

	scheduleMap, err := apc.GetScheduleMap()
	if err != nil {
		return err
	}

	for _, item := range apc.assessments {
		if schedule, ok := scheduleMap[item.ScheduleID]; ok {
			var remainingTime int64
			if schedule.DueAt != 0 {
				remainingTime = schedule.DueAt - time.Now().Unix()
			} else {
				remainingTime = time.Unix(item.CreateAt, 0).Add(config.Get().Assessment.DefaultRemainingTime).Unix() - time.Now().Unix()
			}
			if remainingTime < 0 {
				remainingTime = 0
			}
			apc.assRemainingTimeMap[item.ID] = remainingTime
		}
	}

	return nil
}

func (apc *AssessmentPageComponent) MatchCompleteRate() error {
	if len(apc.assCompleteRateMap) > 0 {
		return nil
	}

	assessmentUserMap, err := apc.GetAssessmentUserMap()
	if err != nil {
		return err
	}

	roomDataMap, err := apc.GetRoomData()
	if err != nil {
		return err
	}

	for _, item := range apc.assessments {
		if roomData, ok := roomDataMap[item.ScheduleID]; ok {
			apc.assCompleteRateMap[item.ID] = getAssessmentLiveRoom().calcRoomCompleteRate(roomData, len(assessmentUserMap[item.ID])-len(apc.assTeacherMap[item.ID]))
		}
	}

	return nil
}

func (apc *AssessmentPageComponent) ConvertPageReply(configs []AssessmentConfigFunc) ([]*v2.AssessmentQueryReply, error) {
	ctx := apc.ctx

	result := make([]*v2.AssessmentQueryReply, len(apc.assessments))

	for _, cfg := range configs {
		err := cfg()
		if err != nil {
			return nil, err
		}
	}

	for i, item := range apc.assessments {
		replyItem := &v2.AssessmentQueryReply{
			ID:         item.ID,
			Title:      item.Title,
			ClassEndAt: item.ClassEndAt,
			CompleteAt: item.CompleteAt,
			Status:     item.Status,
		}
		result[i] = replyItem

		schedule, ok := apc.assScheduleMap[item.ID]
		if !ok {
			log.Warn(ctx, "not found assessment schedule", log.Any("assScheduleMap", apc.assScheduleMap), log.Any("assessmentItem", item))
			continue
		}
		if lessPlanItem, ok := apc.assLessPlanMap[item.ID]; ok && lessPlanItem != nil {
			replyItem.LessonPlan = &entity.IDName{
				ID:   lessPlanItem.ID,
				Name: lessPlanItem.Name,
			}
		}

		replyItem.Program = apc.assProgramMap[item.ID]
		replyItem.Subjects = apc.assSubjectMap[item.ID]
		replyItem.Teachers = apc.assTeacherMap[item.ID]
		replyItem.DueAt = schedule.DueAt
		replyItem.ClassInfo = apc.assClassMap[item.ID]
		replyItem.RemainingTime = apc.assRemainingTimeMap[item.ID]
		replyItem.CompleteRate = apc.assCompleteRateMap[item.ID]
	}

	return result, nil
}
