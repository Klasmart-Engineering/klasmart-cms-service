package model

import (
	"context"
	"fmt"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da/assessmentV2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	v2 "gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

type AssessmentDetailComponent struct {
	ctx context.Context
	op  *entity.Operator

	assessment *v2.Assessment

	apc *AssessmentPageComponent

	roomData *RoomInfo

	contentMapFromSchedule   map[string]*v2.AssessmentContentView // key: contentID
	contentMapFromLiveRoom   map[string]*RoomContent              // key: contentID
	contentMapFromAssessment map[string]*v2.AssessmentContent     // key: contentID

	userMapFromRoomMap       map[string]*RoomUserInfo              // key: userID
	outcomeMapFromContent    map[string]*v2.AssessmentOutcomeReply // key: outcomeID
	outcomeMapFromAssessment map[string]*v2.AssessmentUserOutcome  // key: AssessmentUserID+AssessmentContentID+OutcomeID
	commentResultMap         map[string]string                     //key:userID

	contentsFromSchedule []*v2.AssessmentContentView

	contents []*v2.AssessmentContentReply
	outcomes []*v2.AssessmentOutcomeReply
	students []*v2.AssessmentStudentReply
}

func NewAssessmentDetailComponent(ctx context.Context, op *entity.Operator, assessment *v2.Assessment) *AssessmentDetailComponent {
	return &AssessmentDetailComponent{
		ctx:        ctx,
		op:         op,
		assessment: assessment,

		apc: NewPageComponent(ctx, op, []*v2.Assessment{assessment}),

		roomData: new(RoomInfo),

		contentMapFromSchedule:   make(map[string]*v2.AssessmentContentView),
		contentMapFromLiveRoom:   make(map[string]*RoomContent),
		contentMapFromAssessment: make(map[string]*v2.AssessmentContent),

		userMapFromRoomMap:       make(map[string]*RoomUserInfo),
		outcomeMapFromContent:    make(map[string]*v2.AssessmentOutcomeReply),
		outcomeMapFromAssessment: make(map[string]*v2.AssessmentUserOutcome),
		commentResultMap:         make(map[string]string),

		contents: make([]*v2.AssessmentContentReply, 0),
		outcomes: make([]*v2.AssessmentOutcomeReply, 0),
		students: make([]*v2.AssessmentStudentReply, 0),
	}
}

func (adc *AssessmentDetailComponent) getKey(value []string) string {
	return strings.Join(value, "_")
}

func (adc *AssessmentDetailComponent) isNeedConvertLatestContent() (bool, error) {
	scheduleMap, err := adc.apc.GetScheduleMap()
	if err != nil {
		return false, err
	}

	schedule, ok := scheduleMap[adc.assessment.ScheduleID]
	if !ok {
		return false, constant.ErrRecordNotFound
	}

	if adc.assessment.MigrateFlag == constant.AssessmentHistoryFlag || !schedule.IsLockedLessonPlan() {
		return true, nil
	}

	return false, nil
}

func (adc *AssessmentDetailComponent) GetRoomData() (*RoomInfo, error) {
	if adc.roomData.Initialized {
		return adc.roomData, nil
	}

	ctx := adc.ctx
	op := adc.op
	adc.roomData.Initialized = true

	roomData, err := getAssessmentLiveRoom().getRoomResultInfo(ctx, op, adc.assessment.ScheduleID)
	if err != nil {
		return nil, err
	}

	adc.roomData = roomData

	return adc.roomData, nil
}

func (adc *AssessmentDetailComponent) getContentOutcomeIDsMap(contentIDs []string) (map[string][]string, error) {
	ctx := adc.ctx
	op := adc.op

	contentIDs = utils.SliceDeduplication(contentIDs)

	contents, err := GetContentModel().GetContentByIDList(ctx, dbo.MustGetDB(ctx), contentIDs, op)
	if err != nil {
		log.Error(ctx, "toViews: GetContentModel().GetContentByIDList: get failed",
			log.Err(err),
			log.Strings("lesson_plan_ids", contentIDs),
		)
		return nil, err
	}
	result := make(map[string][]string, len(contents))
	for _, item := range contents {
		result[item.ID] = item.Outcomes
	}

	return result, nil
}

func (adc *AssessmentDetailComponent) getScheduleLockedContents(schedule *entity.Schedule) error {
	contentIDs := make([]string, 0)
	contentIDs = append(contentIDs, schedule.LiveLessonPlan.LessonPlanID)
	for _, materialItem := range schedule.LiveLessonPlan.LessonMaterials {
		contentIDs = append(contentIDs, materialItem.LessonMaterialID)
	}

	contentOutcomeIDsMap, err := adc.getContentOutcomeIDsMap(contentIDs)
	if err != nil {
		return err
	}

	latestContentIDMap, err := adc.apc.GetLatestContentIDMapByIDList(contentIDs)

	liveLessonPlan := schedule.LiveLessonPlan

	lessPlan := &v2.AssessmentContentView{
		ID:          liveLessonPlan.LessonPlanID,
		Name:        liveLessonPlan.LessonPlanName,
		ContentType: v2.AssessmentContentTypeLessonPlan,
		OutcomeIDs:  contentOutcomeIDsMap[liveLessonPlan.LessonPlanID],
		LatestID:    latestContentIDMap[liveLessonPlan.LessonPlanID],
	}

	adc.contentMapFromSchedule[liveLessonPlan.LessonPlanID] = lessPlan
	adc.contentsFromSchedule = append(adc.contentsFromSchedule, lessPlan)

	for _, item := range liveLessonPlan.LessonMaterials {
		materialItem := &v2.AssessmentContentView{
			ID:          item.LessonMaterialID,
			Name:        item.LessonMaterialName,
			ContentType: v2.AssessmentContentTypeLessonMaterial,
			OutcomeIDs:  contentOutcomeIDsMap[item.LessonMaterialID],
			LatestID:    latestContentIDMap[item.LessonMaterialID],
		}
		adc.contentMapFromSchedule[liveLessonPlan.LessonPlanID] = materialItem
		adc.contentsFromSchedule = append(adc.contentsFromSchedule, materialItem)
	}

	return nil
}

func (adc *AssessmentDetailComponent) GetScheduleLockedContents(schedule *entity.Schedule) ([]*v2.AssessmentContentView, error) {
	if !schedule.IsLockedLessonPlan() {
		return nil, constant.ErrInvalidArgs
	}
	err := adc.getScheduleLockedContents(schedule)
	if err != nil {
		return nil, err
	}

	return adc.contentsFromSchedule, nil
}

func (adc *AssessmentDetailComponent) getLatestContents(schedule *entity.Schedule) error {
	ctx := adc.ctx
	op := adc.op

	latestLessPlanIDs, err := GetContentModel().GetLatestContentIDByIDList(ctx, dbo.MustGetDB(ctx), []string{schedule.LessonPlanID})
	if err != nil {
		return err
	}
	if len(latestLessPlanIDs) <= 0 {
		return constant.ErrRecordNotFound
	}

	latestLessPlan, err := GetContentModel().GetContentNameByID(ctx, dbo.MustGetDB(ctx), latestLessPlanIDs[0])
	if err != nil {
		return err
	}

	subContents, err := GetContentModel().GetContentSubContentsByID(ctx, dbo.MustGetDB(ctx), latestLessPlan.ID, op, true)
	if err != nil {
		return err
	}

	lessPlan := &v2.AssessmentContentView{
		ID:          latestLessPlan.ID,
		Name:        latestLessPlan.Name,
		ContentType: v2.AssessmentContentTypeLessonPlan,
		OutcomeIDs:  latestLessPlan.OutcomeIDs,
		LatestID:    latestLessPlan.ID,
	}
	adc.contentMapFromSchedule[latestLessPlan.ID] = lessPlan
	adc.contentsFromSchedule = append(adc.contentsFromSchedule, lessPlan)

	for _, item := range subContents {
		subContentItem := &v2.AssessmentContentView{
			ID:          item.ID,
			Name:        item.Name,
			ContentType: v2.AssessmentContentTypeLessonMaterial,
			OutcomeIDs:  item.OutcomeIDs,
			LatestID:    item.ID,
		}
		adc.contentMapFromSchedule[item.ID] = subContentItem
		adc.contentsFromSchedule = append(adc.contentsFromSchedule, subContentItem)
	}

	return nil
}

//func (adc *SingleAssessmentComponent) GetContentMapFromSchedule() (map[string]*v2.AssessmentContentView, error) {
//	if _, ok := adc.contentMapFromSchedule[constant.AssessmentInitializedKey]; ok {
//		return adc.contentMapFromSchedule, nil
//	}
//
//	scheduleMap, err := adc.apc.GetScheduleMap()
//	if err != nil {
//		return nil, err
//	}
//
//	schedule, ok := scheduleMap[adc.assessment.ScheduleID]
//	if !ok {
//		return nil, constant.ErrRecordNotFound
//	}
//
//	if schedule.AnyoneAttemptedLive() {
//		if _, err := adc.getLockContents(schedule); err != nil {
//			return nil, err
//		}
//	} else {
//		if _, err := adc.getUnlockContents(schedule); err != nil {
//			return nil, err
//		}
//	}
//
//	adc.contentMapFromSchedule[constant.AssessmentInitializedKey] = new(v2.AssessmentContentView)
//
//	return adc.contentMapFromSchedule, nil
//}

func (adc *AssessmentDetailComponent) GetContentsFromSchedule() ([]*v2.AssessmentContentView, error) {
	if adc.contentsFromSchedule != nil {
		return adc.contentsFromSchedule, nil
	}

	scheduleMap, err := adc.apc.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	schedule, ok := scheduleMap[adc.assessment.ScheduleID]
	if !ok {
		return nil, constant.ErrRecordNotFound
	}

	if ok, _ := adc.isNeedConvertLatestContent(); ok {
		if err := adc.getLatestContents(schedule); err != nil {
			return nil, err
		}
	} else {
		if err := adc.getScheduleLockedContents(schedule); err != nil {
			return nil, err
		}
	}

	return adc.contentsFromSchedule, nil
}

func (adc *AssessmentDetailComponent) GetContentMapFromLiveRoom() (map[string]*RoomContent, error) {
	ctx := adc.ctx
	//op := adc.op

	if _, ok := adc.contentMapFromLiveRoom[constant.AssessmentInitializedKey]; ok {
		return adc.contentMapFromLiveRoom, nil
	}

	roomInfo, err := adc.GetRoomData()
	if err != nil {
		return nil, err
	}

	if ok, _ := adc.isNeedConvertLatestContent(); ok {
		oldContentIDs := make([]string, len(roomInfo.Contents))
		oldContentMap := make(map[string]*RoomContent)
		for i, item := range roomInfo.Contents {
			oldContentIDs[i] = item.MaterialID
			oldContentMap[item.MaterialID] = item
		}

		latestContentIDMap, err := adc.apc.GetLatestContentIDMapByIDList(oldContentIDs)
		if err != nil {
			log.Error(ctx, "GetLatestContentIDMapByIDList error", log.Err(err), log.Strings("oldContentIDs", oldContentIDs))
			return nil, err
		}

		adc.contentMapFromLiveRoom = make(map[string]*RoomContent, len(roomInfo.Contents))
		for _, item := range roomInfo.Contents {
			adc.contentMapFromLiveRoom[latestContentIDMap[item.MaterialID]] = item
		}
	} else {
		adc.contentMapFromLiveRoom = make(map[string]*RoomContent, len(roomInfo.Contents))
		for _, item := range roomInfo.Contents {
			adc.contentMapFromLiveRoom[item.MaterialID] = item
		}
	}

	adc.contentMapFromLiveRoom[constant.AssessmentInitializedKey] = new(RoomContent)

	return adc.contentMapFromLiveRoom, nil
}

func (adc *AssessmentDetailComponent) GetUserMapFromLiveRoom() (map[string]*RoomUserInfo, error) {
	//ctx := adc.ctx
	//op := adc.op

	if _, ok := adc.userMapFromRoomMap[constant.AssessmentInitializedKey]; ok {
		return adc.userMapFromRoomMap, nil
	}

	roomInfo, err := adc.GetRoomData()
	if err != nil {
		return nil, err
	}

	adc.userMapFromRoomMap = make(map[string]*RoomUserInfo, len(roomInfo.UserRoomInfo))
	for _, item := range roomInfo.UserRoomInfo {
		adc.userMapFromRoomMap[item.UserID] = item
	}

	adc.userMapFromRoomMap[constant.AssessmentInitializedKey] = new(RoomUserInfo)

	return adc.userMapFromRoomMap, nil
}

func (adc *AssessmentDetailComponent) GetOutcomeMap() (map[string]*v2.AssessmentOutcomeReply, error) {
	if _, ok := adc.outcomeMapFromContent[constant.AssessmentInitializedKey]; ok {
		return adc.outcomeMapFromContent, nil
	}

	ctx := adc.ctx
	op := adc.op

	contents, err := adc.GetContentsFromSchedule()
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

	adc.outcomeMapFromContent = make(map[string]*v2.AssessmentOutcomeReply, len(outcomes))

	for _, item := range outcomes {
		adc.outcomeMapFromContent[item.ID] = &v2.AssessmentOutcomeReply{
			OutcomeID:          item.ID,
			OutcomeName:        item.Name,
			AssignedTo:         nil,
			Assumed:            item.Assumed,
			AssignedToLessPlan: false,
			AssignedToMaterial: false,
		}
	}

	for _, materialItem := range contents {
		for _, outcomeID := range materialItem.OutcomeIDs {
			if outcomeItem, ok := adc.outcomeMapFromContent[outcomeID]; ok {
				if materialItem.ContentType == v2.AssessmentContentTypeLessonPlan {
					outcomeItem.AssignedToLessPlan = true
				}
				if materialItem.ContentType == v2.AssessmentContentTypeLessonMaterial {
					outcomeItem.AssignedToMaterial = true
				}
			}
		}
	}

	for _, outcomeItem := range adc.outcomeMapFromContent {
		if outcomeItem.AssignedToLessPlan {
			outcomeItem.AssignedTo = append(outcomeItem.AssignedTo, v2.AssessmentOutcomeAssignTypeLessonPlan)
		}
		if outcomeItem.AssignedToMaterial {
			outcomeItem.AssignedTo = append(outcomeItem.AssignedTo, v2.AssessmentOutcomeAssignTypeLessonMaterial)
		}
	}

	adc.outcomeMapFromContent[constant.AssessmentInitializedKey] = new(v2.AssessmentOutcomeReply)

	return adc.outcomeMapFromContent, nil
}

func (adc *AssessmentDetailComponent) GetOutcomeFromAssessment() (map[string]*v2.AssessmentUserOutcome, error) {
	ctx := adc.ctx

	assessmentUserMap, err := adc.apc.GetAssessmentUserMap()
	if err != nil {
		return nil, err
	}
	assessmentUsers, ok := assessmentUserMap[adc.assessment.ID]
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
	adc.outcomeMapFromAssessment = make(map[string]*v2.AssessmentUserOutcome, len(userOutcomes))
	for _, item := range userOutcomes {
		key := adc.getKey([]string{
			item.AssessmentUserID,
			item.AssessmentContentID,
			item.OutcomeID,
		})
		adc.outcomeMapFromAssessment[key] = item
	}

	return adc.outcomeMapFromAssessment, nil
}

func (adc *AssessmentDetailComponent) GetCommentResultMap() (map[string]string, error) {
	if _, ok := adc.commentResultMap[constant.AssessmentInitializedKey]; ok {
		return adc.commentResultMap, nil
	}

	commentResults, err := getAssessmentLiveRoom().batchGetRoomCommentMap(adc.ctx, adc.op, []string{adc.assessment.ScheduleID})
	if err != nil {
		log.Error(adc.ctx, "get assessment comment from live room error", log.Err(err), log.String("scheduleID", adc.assessment.ScheduleID))
	} else {
		if commentItem, ok := commentResults[adc.assessment.ScheduleID]; ok && commentItem != nil {
			adc.commentResultMap = commentItem
		}
	}

	adc.commentResultMap[constant.AssessmentInitializedKey] = ""

	return adc.commentResultMap, nil
}

func (adc *AssessmentDetailComponent) GetAssessmentContentMap() (map[string]*v2.AssessmentContent, error) {
	ctx := adc.ctx

	assessmentContentMap, err := adc.apc.GetAssessmentContentMap()
	if err != nil {
		return nil, err
	}

	assessmentContents := assessmentContentMap[adc.assessment.ID]

	if ok, _ := adc.isNeedConvertLatestContent(); ok {
		oldContentIDs := make([]string, len(assessmentContents))
		assessmentContentMap := make(map[string]*v2.AssessmentContent)
		for i, item := range assessmentContents {
			oldContentIDs[i] = item.ContentID
			assessmentContentMap[item.ContentID] = item
		}

		latestContentIDMap, err := adc.apc.GetLatestContentIDMapByIDList(oldContentIDs)
		if err != nil {
			log.Error(ctx, "GetLatestContentIDMapByIDList error", log.Err(err), log.Strings("oldContentIDs", oldContentIDs))
			return nil, err
		}

		adc.contentMapFromAssessment = make(map[string]*v2.AssessmentContent, len(latestContentIDMap))
		for newID, oldID := range latestContentIDMap {
			adc.contentMapFromAssessment[newID] = assessmentContentMap[oldID]
		}
	} else {
		adc.contentMapFromAssessment = make(map[string]*v2.AssessmentContent, len(assessmentContents))
		for _, item := range assessmentContents {
			adc.contentMapFromAssessment[item.ContentID] = item
		}
	}

	return adc.contentMapFromAssessment, nil
}

func (adc *AssessmentDetailComponent) GetAssessmentUserWithUserIDAndUserTypeMap() (map[string]*v2.AssessmentUser, error) {
	ctx := adc.ctx

	assessmentUserMap, err := adc.apc.GetAssessmentUserMap()
	if err != nil {
		return nil, err
	}

	assessmentUsers, ok := assessmentUserMap[adc.assessment.ID]
	if !ok {
		log.Error(ctx, "not found assessment users", log.Any("assessment", adc.assessment), log.Any("assessmentUserMap", assessmentUserMap))
		return nil, constant.ErrRecordNotFound
	}

	result := make(map[string]*v2.AssessmentUser, len(assessmentUsers))
	for _, item := range assessmentUsers {
		result[adc.getKey([]string{item.UserID, item.UserType.String()})] = item
	}

	return result, nil
}

func (adc *AssessmentDetailComponent) MatchContentsNotContainsRoomInfo() error {
	if len(adc.contents) > 0 {
		return nil
	}

	libraryContents, err := adc.GetContentsFromSchedule()
	if err != nil {
		return err
	}

	assessmentContentMap, err := adc.GetAssessmentContentMap()
	if err != nil {
		return err
	}

	index := 0
	for _, item := range libraryContents {
		contentReplyItem := &v2.AssessmentContentReply{
			Number:          "0",
			ParentID:        "",
			ContentID:       item.ID,
			ContentName:     item.Name,
			Status:          v2.AssessmentContentStatusCovered,
			ContentType:     item.ContentType,
			FileType:        v2.AssessmentFileTypeNotChildSubContainer,
			MaxScore:        0,
			ReviewerComment: "",
			OutcomeIDs:      item.OutcomeIDs,
		}

		if item.ContentType == v2.AssessmentContentTypeLessonPlan {
			contentReplyItem.FileType = v2.AssessmentFileTypeHasChildContainer
			adc.contents = append(adc.contents, contentReplyItem)
			continue
		}

		index++
		contentReplyItem.Number = fmt.Sprintf("%d", index)

		if assessmentContentItem, ok := assessmentContentMap[item.ID]; ok {
			contentReplyItem.ReviewerComment = assessmentContentItem.ReviewerComment
			contentReplyItem.Status = assessmentContentItem.Status
		}

		adc.contents = append(adc.contents, contentReplyItem)
	}

	return nil
}

func (adc *AssessmentDetailComponent) MatchContentsContainsRoomInfo() error {
	if len(adc.contents) > 0 {
		return nil
	}

	libraryContents, err := adc.GetContentsFromSchedule()
	if err != nil {
		return err
	}

	assessmentContentMap, err := adc.GetAssessmentContentMap()
	if err != nil {
		return err
	}

	roomContentMap, err := adc.GetContentMapFromLiveRoom()
	if err != nil {
		return err
	}

	index := 0
	for _, item := range libraryContents {
		contentReplyItem := &v2.AssessmentContentReply{
			Number:               "0",
			ParentID:             "",
			ContentID:            item.ID,
			ContentName:          item.Name,
			Status:               v2.AssessmentContentStatusCovered,
			ContentType:          item.ContentType,
			FileType:             v2.AssessmentFileTypeNotChildSubContainer,
			MaxScore:             0,
			ReviewerComment:      "",
			OutcomeIDs:           item.OutcomeIDs,
			RoomProvideContentID: "",
		}

		if item.ContentType == v2.AssessmentContentTypeLessonPlan {
			contentReplyItem.FileType = v2.AssessmentFileTypeHasChildContainer
			adc.contents = append(adc.contents, contentReplyItem)
			continue
		}

		index++
		contentReplyItem.Number = fmt.Sprintf("%d", index)

		if assessmentContentItem, ok := assessmentContentMap[item.ID]; ok {
			contentReplyItem.ReviewerComment = assessmentContentItem.ReviewerComment
			contentReplyItem.Status = assessmentContentItem.Status
		}

		if roomContentItem, ok := roomContentMap[item.ID]; ok {
			contentReplyItem.ContentSubtype = roomContentItem.SubContentType
			contentReplyItem.H5PID = roomContentItem.H5PID
			contentReplyItem.MaxScore = roomContentItem.MaxScore
			contentReplyItem.RoomProvideContentID = roomContentItem.ID

			if roomContentItem.FileType == external.FileTypeH5P {
				if canScoringMap[roomContentItem.SubContentType] {
					contentReplyItem.FileType = v2.AssessmentFileTypeSupportScoreStandAlone
				} else {
					contentReplyItem.FileType = v2.AssessmentFileTypeNotSupportScoreStandAlone
				}
			} else {
				contentReplyItem.FileType = v2.AssessmentFileTypeNotChildSubContainer
			}

			if len(roomContentItem.Children) > 0 {
				contentReplyItem.FileType = v2.AssessmentFileTypeHasChildContainer
				adc.contents = append(adc.contents, contentReplyItem)

				for i, child := range roomContentItem.Children {
					adc.appendContent(child, item, &adc.contents, contentReplyItem.Number, i+1)
				}
			} else {
				adc.contents = append(adc.contents, contentReplyItem)
			}
		} else {
			adc.contents = append(adc.contents, contentReplyItem)
		}
	}
	return nil
}

func (adc *AssessmentDetailComponent) MatchOutcome() error {
	outcomeMap, err := adc.GetOutcomeMap()
	if err != nil {
		return err
	}

	adc.outcomes = make([]*v2.AssessmentOutcomeReply, 0, len(outcomeMap))
	for _, item := range outcomeMap {
		adc.outcomes = append(adc.outcomes, item)
	}

	return nil
}

func (adc *AssessmentDetailComponent) MatchStudentNotContainsRoomInfo() error {
	userMap, err := adc.apc.GetUserMap()
	if err != nil {
		return err
	}

	assessmentUserMap, err := adc.apc.GetAssessmentUserMap()
	if err != nil {
		return err
	}

	assessmentUsers, ok := assessmentUserMap[adc.assessment.ID]
	if !ok {
		return constant.ErrRecordNotFound
	}

	err = adc.MatchContentsNotContainsRoomInfo()
	if err != nil {
		return err
	}

	contentMapFromAssessment, err := adc.GetAssessmentContentMap()
	if err != nil {
		return err
	}

	assessmentOutcomeMap, err := adc.GetOutcomeFromAssessment()
	if err != nil {
		return err
	}

	allOutcomeMap, err := adc.GetOutcomeMap()
	if err != nil {
		return err
	}

	for _, item := range assessmentUsers {
		if item.UserType == v2.AssessmentUserTypeTeacher {
			continue
		}

		studentInfo, ok := userMap[item.UserID]
		if !ok {
			continue
		}

		if adc.assessment.AssessmentType == v2.AssessmentTypeOnlineClass && item.StatusByUser == v2.AssessmentUserStatusNotParticipate {
			continue
		}

		studentReply := &v2.AssessmentStudentReply{
			StudentID:   item.UserID,
			StudentName: studentInfo.Name,
			Status:      item.StatusByUser,
			Results:     nil,
		}

		for _, content := range adc.contents {
			resultReply := &v2.AssessmentStudentResultReply{
				ContentID: content.ContentID,
			}

			userOutcomeReply := make([]*v2.AssessmentStudentResultOutcomeReply, 0)
			for _, outcomeID := range content.OutcomeIDs {
				var userOutcome *v2.AssessmentUserOutcome
				if assessmentContent, ok := contentMapFromAssessment[content.ContentID]; ok {
					key := adc.getKey([]string{
						item.ID,
						assessmentContent.ID,
						outcomeID,
					})
					userOutcome = assessmentOutcomeMap[key]
				}
				userOutcomeReplyItem := &v2.AssessmentStudentResultOutcomeReply{
					OutcomeID: outcomeID,
				}
				if userOutcome != nil {
					userOutcomeReplyItem.Status = userOutcome.Status
				} else {
					if outcomeInfo, ok := allOutcomeMap[outcomeID]; ok && outcomeInfo.Assumed {
						userOutcomeReplyItem.Status = v2.AssessmentUserOutcomeStatusAchieved
					} else {
						userOutcomeReplyItem.Status = v2.AssessmentUserOutcomeStatusUnknown
					}
				}

				userOutcomeReply = append(userOutcomeReply, userOutcomeReplyItem)
			}
			resultReply.Outcomes = userOutcomeReply

			studentReply.Results = append(studentReply.Results, resultReply)
		}

		adc.students = append(adc.students, studentReply)
	}

	return nil
}

func (adc *AssessmentDetailComponent) MatchStudentContainsRoomInfo() error {
	assessmentUserMap, err := adc.apc.GetAssessmentUserMap()
	if err != nil {
		return err
	}

	assessmentUsers, ok := assessmentUserMap[adc.assessment.ID]
	if !ok {
		return constant.ErrRecordNotFound
	}

	err = adc.MatchContentsContainsRoomInfo()
	if err != nil {
		return err
	}

	commentResultMap, err := adc.GetCommentResultMap()
	if err != nil {
		return err
	}

	assessmentOutcomeMap, err := adc.GetOutcomeFromAssessment()
	if err != nil {
		return err
	}

	userMapFromRoomMap, err := adc.GetUserMapFromLiveRoom()
	if err != nil {
		return err
	}

	userMap, err := adc.apc.GetUserMap()
	if err != nil {
		return err
	}

	roomUserResultMap := make(map[string]*RoomUserResults)
	for _, item := range userMapFromRoomMap {
		for _, resultItem := range item.Results {
			key := adc.getKey([]string{
				item.UserID,
				resultItem.RoomContentID,
			})
			roomUserResultMap[key] = resultItem
		}
	}

	for _, item := range assessmentUsers {
		if item.UserType == v2.AssessmentUserTypeTeacher {
			continue
		}

		studentInfo, ok := userMap[item.UserID]
		if !ok {
			continue
		}

		studentReply := &v2.AssessmentStudentReply{
			StudentID:   item.UserID,
			StudentName: studentInfo.Name,
			Status:      item.StatusByUser,
			Results:     nil,
		}
		studentReply.ReviewerComment = commentResultMap[item.UserID]

		for _, content := range adc.contents {
			resultReply := &v2.AssessmentStudentResultReply{
				ContentID: content.ContentID,
			}

			userOutcomeReply := make([]*v2.AssessmentStudentResultOutcomeReply, 0)
			for _, outcomeID := range content.OutcomeIDs {
				var userOutcome *v2.AssessmentUserOutcome
				if assessmentContent, ok := adc.contentMapFromAssessment[content.ContentID]; ok {
					key := adc.getKey([]string{
						item.ID,
						assessmentContent.ID,
						outcomeID,
					})
					userOutcome = assessmentOutcomeMap[key]
				}
				userOutcomeReplyItem := &v2.AssessmentStudentResultOutcomeReply{
					OutcomeID: outcomeID,
				}
				if userOutcome != nil {
					userOutcomeReplyItem.Status = userOutcome.Status
				} else {
					if outcomeInfo, ok := adc.outcomeMapFromContent[outcomeID]; ok && outcomeInfo.Assumed {
						userOutcomeReplyItem.Status = v2.AssessmentUserOutcomeStatusAchieved
					} else {
						userOutcomeReplyItem.Status = v2.AssessmentUserOutcomeStatusUnknown
					}
				}

				userOutcomeReply = append(userOutcomeReply, userOutcomeReplyItem)
			}
			resultReply.Outcomes = userOutcomeReply

			//if roomContent, ok := adc.contentMapFromLiveRoom[content.ContentID]; ok {
			//	roomKey := adc.getKey([]string{
			//		item.UserID,
			//		roomContent.ID,
			//	})
			//	if roomResultItem, ok := roomUserResultMap[roomKey]; ok {
			//		resultReply.Answer = roomResultItem.Answer
			//		resultReply.Score = roomResultItem.Score
			//		resultReply.Attempted = roomResultItem.Seen
			//	}
			//} else {
			//
			//}

			roomKey := adc.getKey([]string{
				item.UserID,
				content.RoomProvideContentID,
			})
			if roomResultItem, ok := roomUserResultMap[roomKey]; ok {
				resultReply.Answer = roomResultItem.Answer
				resultReply.Score = roomResultItem.Score
				resultReply.Attempted = roomResultItem.Seen
			}

			studentReply.Results = append(studentReply.Results, resultReply)
		}

		adc.students = append(adc.students, studentReply)
	}

	log.Debug(adc.ctx, "MatchStudentContainsRoomInfo data",
		log.Any("roomUserResultMap", roomUserResultMap),
		log.Any("adc.contents", adc.contents),
		log.Any("adc.userMapFromRoomMap", adc.userMapFromRoomMap))

	return nil
}

func (adc *AssessmentDetailComponent) appendContent(roomContent *RoomContent, materialItem *v2.AssessmentContentView, result *[]*v2.AssessmentContentReply, prefix string, index int) {
	replyItem := &v2.AssessmentContentReply{
		Number:               fmt.Sprintf("%s-%d", prefix, index),
		ParentID:             materialItem.ID,
		ContentID:            roomContent.ID,
		ContentName:          materialItem.Name,
		ReviewerComment:      "",
		Status:               v2.AssessmentContentStatusCovered,
		OutcomeIDs:           materialItem.OutcomeIDs,
		ContentType:          v2.AssessmentContentTypeUnknown,
		ContentSubtype:       roomContent.SubContentType,
		FileType:             v2.AssessmentFileTypeNotUnknown,
		MaxScore:             roomContent.MaxScore,
		H5PID:                roomContent.H5PID,
		RoomProvideContentID: roomContent.ID,
		//LatestID:       materialItem.LatestID,
	}

	if roomContent.FileType == external.FileTypeH5P {
		if canScoringMap[roomContent.SubContentType] {
			replyItem.FileType = v2.AssessmentFileTypeSupportScoreStandAlone
		} else {
			replyItem.FileType = v2.AssessmentFileTypeNotSupportScoreStandAlone
		}
	}

	*result = append(*result, replyItem)
	for i, item := range roomContent.Children {
		adc.appendContent(item, materialItem, result, replyItem.Number, i+1)
	}
}

func (adc *AssessmentDetailComponent) ConvertDetailReply(configs []AssessmentConfigFunc) (*v2.AssessmentDetailReply, error) {
	for _, cfg := range configs {
		err := cfg()
		if err != nil {
			return nil, err
		}
	}

	result := &v2.AssessmentDetailReply{
		ID:          adc.assessment.ID,
		Title:       adc.assessment.Title,
		Status:      adc.assessment.Status,
		RoomID:      adc.assessment.ScheduleID,
		ClassEndAt:  adc.assessment.ClassEndAt,
		ClassLength: adc.assessment.ClassLength,
		CompleteAt:  adc.assessment.CompleteAt,
		Outcomes:    nil,
		Contents:    nil,
		Students:    nil,
	}

	schedule, ok := adc.apc.allScheduleMap[adc.assessment.ScheduleID]
	if !ok {
		return nil, constant.ErrRecordNotFound
	}

	result.Class = adc.apc.assClassMap[adc.assessment.ID]
	result.Teachers = adc.apc.assTeacherMap[adc.assessment.ID]
	result.Program = adc.apc.assProgramMap[adc.assessment.ID]
	result.Subjects = adc.apc.assSubjectMap[adc.assessment.ID]
	result.ScheduleTitle = schedule.Title
	result.ScheduleDueAt = schedule.DueAt
	result.RemainingTime = adc.apc.assRemainingTimeMap[adc.assessment.ID]
	result.Outcomes = adc.outcomes
	result.Contents = adc.contents
	result.Students = adc.students

	return result, nil
}
