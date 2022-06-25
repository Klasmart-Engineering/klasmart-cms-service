package model

import (
	"context"
	"database/sql"
	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/da/assessmentV2"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
	"github.com/KL-Engineering/kidsloop-cms-service/utils/errgroup"
	"strings"
	"time"
)

type assessmentInitFunc func() error

type assessmentInitLevel int

const (
	assessmentInitLevel1 assessmentInitLevel = iota
	assessmentInitLevel2
	assessmentInitLevel3
	assessmentInitLevel4
)

type AssessmentDetailInit struct {
	ctx           context.Context
	op            *entity.Operator
	assessment    *v2.Assessment
	detailInitMap map[assessmentInitLevel][]assessmentInitFunc

	lai *AssessmentListInit

	schedule          *entity.Schedule           // scheduleID
	scheduleRelations []*entity.ScheduleRelation // scheduleID
	assessmentUsers   []*v2.AssessmentUser

	lessPlanMap *v2.AssessmentContentView // lessPlanID
	//teacher           []*entity.IDName                          // userID
	program              *entity.IDName                            // programID
	subjects             []*entity.IDName                          // subjectID
	class                *entity.IDName                            // classID
	reviewerFeedbackMap  map[string]*v2.AssessmentReviewerFeedback // assessmentUserID
	liveRoom             *external.RoomInfo                        // roomID
	scheduleStuReviewMap map[string]*entity.ScheduleReview         // key:ScheduleID,StudentID

	outcomeMapFromContent map[string]*entity.Outcome // key: outcomeID
	contentsFromSchedule  []*v2.AssessmentContentView
	//lockedContentsFromSchedule []*v2.AssessmentContentView
	contentMapFromAssessment map[string]*v2.AssessmentContent // key:contentID
	contentMapFromLiveRoom   map[string]*RoomContentTree      // key: contentID
	// key:userID if the data comes from external assessment service
	// key:assessment_user_id if the data comes from reviewerFeedback
	commentResultMap         map[string]string
	outcomeMapFromAssessment map[string]*v2.AssessmentUserOutcome // key: AssessmentUserID+AssessmentContentID+OutcomeID
	outcomesFromSchedule     []*entity.Outcome

	scheduleReviewMap            map[string]*entity.ScheduleReview      // key:StudentID
	contentMapFromScheduleReview map[string]*entity.ContentInfoInternal // key:contentID

	roomUserScoreMap map[string][]*RoomUserScore // key: userID
	roomContentTree  []*RoomContentTree

	assessmentUserIDTypeMap map[string]*v2.AssessmentUser // key: userID+userType
}

func NewAssessmentDetailInit(ctx context.Context, op *entity.Operator, assessment *v2.Assessment) (*AssessmentDetailInit, error) {
	if assessment == nil {
		return nil, constant.ErrRecordNotFound
	}

	lai, err := NewAssessmentListInit(ctx, op, []*v2.Assessment{assessment})
	if err != nil {
		return nil, err
	}

	at := &AssessmentDetailInit{
		ctx:           ctx,
		op:            op,
		assessment:    assessment,
		lai:           lai,
		detailInitMap: make(map[assessmentInitLevel][]assessmentInitFunc),
	}

	at.ConfigAssessmentInitMap()

	return at, nil
}

func (at *AssessmentDetailInit) ConfigAssessmentInitMap() {
	data := make(map[assessmentInitLevel][]assessmentInitFunc)
	data[assessmentInitLevel1] = append(data[assessmentInitLevel1],
		at.initSchedule,
		at.initScheduleRelation,
		at.initScheduleReviewMap,
		at.initAssessmentUsers,
		at.initLiveRoom)

	data[assessmentInitLevel2] = append(data[assessmentInitLevel2],
		at.initProgram,
		at.initSubject,
		at.initClass,
		at.initAssessmentUserWithIDTypeMap,
		at.initContentsFromSchedule,
		at.initAssessmentContentMap,
		at.initLiveRoom,
		at.initReviewerFeedbackMap,
		at.initContentsFromScheduleReview,
		at.initOutcomesFromSchedule)

	data[assessmentInitLevel3] = append(data[assessmentInitLevel3],
		at.initOutcomeMapFromContent,
		at.initRoomData,
		at.initCommentResultMap,
		at.initOutcomeFromAssessment,
		at.initContentMapFromLiveRoom)

	at.detailInitMap = data
}

// Level 1

func (at *AssessmentDetailInit) initSchedule() error {
	if err := at.lai.initScheduleMap(); err != nil {
		return err
	}

	at.schedule = at.lai.scheduleMap[at.assessment.ScheduleID]

	return nil
}

func (at *AssessmentDetailInit) initScheduleRelation() error {
	if err := at.lai.initScheduleRelationMap(); err != nil {
		return err
	}

	at.scheduleRelations = at.lai.scheduleRelationMap[at.assessment.ScheduleID]

	return nil
}

func (at *AssessmentDetailInit) initScheduleReviewMap() error {

	if err := at.lai.initScheduleReviewMap(); err != nil {
		return err
	}

	at.scheduleStuReviewMap = at.lai.scheduleStuReviewMap[at.assessment.ScheduleID]

	return nil
}

func (at *AssessmentDetailInit) initAssessmentUsers() error {
	if err := at.lai.initAssessmentUsers(); err != nil {
		return err
	}
	at.assessmentUsers = at.lai.assessmentUsers

	return nil
}

func (at *AssessmentDetailInit) GetKey(value []string) string {
	return strings.Join(value, "_")
}

func (at *AssessmentDetailInit) initLiveRoom() error {
	ctx := at.ctx
	op := at.op

	if at.assessment.AssessmentType == v2.AssessmentTypeOfflineClass ||
		at.assessment.AssessmentType == v2.AssessmentTypeOfflineStudy {
		at.liveRoom = &external.RoomInfo{}
		return nil
	}

	roomDataMap, err := external.GetAssessmentServiceProvider().Get(ctx, op,
		[]string{at.assessment.ScheduleID},
		external.WithAssessmentGetScore(false),
		external.WithAssessmentGetTeacherComment(true))
	if err != nil {
		log.Warn(ctx, "external service error",
			log.Err(err), log.Any("assessment", at.assessment), log.Any("op", at.op))
		at.liveRoom = &external.RoomInfo{}
	} else {
		if item, ok := roomDataMap[at.assessment.ScheduleID]; ok {
			at.liveRoom = item
		}
	}

	return nil
}

// Level 2
func (at *AssessmentDetailInit) initProgram() error {
	if err := at.lai.initProgramMap(); err != nil {
		return err
	}

	for _, item := range at.lai.programMap {
		at.program = item
		break
	}

	return nil
}

func (at *AssessmentDetailInit) initSubject() error {
	if err := at.lai.initSubjectMap(); err != nil {
		return err
	}
	for _, item := range at.lai.subjectMap {
		at.subjects = append(at.subjects, item)
	}

	return nil
}

func (at *AssessmentDetailInit) initClass() error {
	if err := at.lai.initClassMap(); err != nil {
		return err
	}

	for _, item := range at.lai.classMap {
		at.class = item
		break
	}

	return nil
}

func (at *AssessmentDetailInit) IsNeedConvertLatestContent() (bool, error) {
	ctx := at.ctx

	schedule := at.schedule

	if (at.assessment.MigrateFlag == constant.AssessmentHistoryFlag &&
		at.assessment.Status != v2.AssessmentStatusNotStarted) || !schedule.IsLockedLessonPlan() {
		log.Debug(ctx, "assessment belongs to the migration or schedule can not locked lessPlan", log.Any("assessment", at.assessment))
		return true, nil
	}

	return false, nil
}

func (at *AssessmentDetailInit) initAssessmentUserWithIDTypeMap() error {
	//ctx := at.ctx

	assessmentUsers := at.assessmentUsers

	result := make(map[string]*v2.AssessmentUser, len(assessmentUsers))
	for _, item := range assessmentUsers {
		result[at.GetKey([]string{item.UserID, item.UserType.String()})] = item
	}

	at.assessmentUserIDTypeMap = result

	return nil
}

func (at *AssessmentDetailInit) getLatestContentsFromSchedule(schedule *entity.Schedule) ([]*v2.AssessmentContentView, error) {
	ctx := at.ctx
	op := at.op

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

	latestLessPlan := latestLessPlans[0]
	subContents := subContentsMap[latestLessPlan.ID]

	result := make([]*v2.AssessmentContentView, 0, len(subContents)+1)
	// filling lesson plan
	lessPlan := &v2.AssessmentContentView{
		ID:          latestLessPlan.ID,
		Name:        latestLessPlan.Name,
		ContentType: v2.AssessmentContentTypeLessonPlan,
		OutcomeIDs:  latestLessPlan.OutcomeIDs,
		LatestID:    latestLessPlan.ID,
		FileType:    latestLessPlan.FileType,
	}
	result = append(result, lessPlan)

	// filling material
	dedupMap := make(map[string]struct{})

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
		result = append(result, subContentItem)
		dedupMap[item.ID] = struct{}{}
	}

	// convert content outcome to latest outcome
	err = at.convertContentOutcome(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
func (at *AssessmentDetailInit) convertContentOutcome(contents []*v2.AssessmentContentView) error {
	ctx := at.ctx
	op := at.op

	contentOutcomes, err := assessmentV2.GetAssessmentUserOutcomeDA().GetContentOutcomeByAssessmentID(ctx, at.assessment.ID)
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
func (at *AssessmentDetailInit) getLockedContentsFromSchedule(schedule *entity.Schedule) ([]*v2.AssessmentContentView, error) {
	result, err := at.getLockedContentBySchedule(schedule)
	if err != nil {
		return nil, err
	}

	err = at.convertContentOutcome(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
func (at *AssessmentDetailInit) getLockedContentBySchedule(schedule *entity.Schedule) ([]*v2.AssessmentContentView, error) {
	ctx := at.ctx

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

	dedupMap = make(map[string]struct{})
	for _, materialItem := range schedule.LiveLessonPlan.LessonMaterials {
		if _, ok := dedupMap[materialItem.LessonMaterialID]; ok {
			log.Warn(ctx, "this content already exists", log.Any("dedupMap", dedupMap), log.String("LessonMaterialID", materialItem.LessonMaterialID), log.Any("contentMap", contentMap))
			continue
		}
		dedupMap[materialItem.LessonMaterialID] = struct{}{}

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
func (at *AssessmentDetailInit) initContentsFromSchedule() error {
	at.contentsFromSchedule = make([]*v2.AssessmentContentView, 0)

	if at.assessment.AssessmentType != v2.AssessmentTypeOnlineClass ||
		at.assessment.AssessmentType != v2.AssessmentTypeOfflineClass ||
		at.assessment.AssessmentType != v2.AssessmentTypeOnlineStudy {
		return nil
	}

	schedule := at.schedule
	if schedule == nil {
		return ErrorPreambleDataNotInitialized
	}

	if ok, _ := at.IsNeedConvertLatestContent(); ok {
		result, err := at.getLatestContentsFromSchedule(schedule)
		if err != nil {
			return err
		}
		at.contentsFromSchedule = result
	} else {
		result, err := at.getLockedContentsFromSchedule(schedule)
		if err != nil {
			return err
		}
		at.contentsFromSchedule = result
	}
	return nil
}

func (at *AssessmentDetailInit) initAssessmentContentMap() error {
	ctx := at.ctx

	var assessmentContents []*v2.AssessmentContent
	err := assessmentV2.GetAssessmentContentDA().Query(ctx, &assessmentV2.AssessmentUserCondition{
		AssessmentID: sql.NullString{
			String: at.assessment.ID,
			Valid:  true,
		},
	}, &assessmentContents)
	if err != nil {
		return err
	}

	if ok, _ := at.IsNeedConvertLatestContent(); ok {
		oldContentIDs := make([]string, len(assessmentContents))
		assessmentContentMap := make(map[string]*v2.AssessmentContent)
		for i, item := range assessmentContents {
			oldContentIDs[i] = item.ContentID
			assessmentContentMap[item.ContentID] = item
		}

		latestContentIDMap, err := GetContentModel().GetLatestContentIDMapByIDListInternal(ctx, dbo.MustGetDB(ctx), oldContentIDs)
		if err != nil {
			log.Error(ctx, "GetLatestContentIDMapByIDList error", log.Err(err), log.Strings("oldContentIDs", oldContentIDs))
			return err
		}

		result := make(map[string]*v2.AssessmentContent, len(latestContentIDMap))
		for oldID, newID := range latestContentIDMap {
			result[newID] = assessmentContentMap[oldID]
		}

		at.contentMapFromAssessment = result
	} else {
		result := make(map[string]*v2.AssessmentContent, len(assessmentContents))
		for _, item := range assessmentContents {
			result[item.ContentID] = item
		}

		at.contentMapFromAssessment = result
	}

	return nil
}

func (at *AssessmentDetailInit) initReviewerFeedbackMap() error {
	if err := at.lai.initReviewerFeedbackMap(); err != nil {
		return err
	}

	at.reviewerFeedbackMap = at.lai.reviewerFeedbackMap

	return nil
}

func (at *AssessmentDetailInit) initContentsFromScheduleReview() error {
	ctx := at.ctx
	//op := at.op

	contentIDs := make([]string, 0)
	dedupContentID := make(map[string]struct{})
	scheduleReviewMap := at.scheduleReviewMap
	if scheduleReviewMap == nil {
		log.Error(ctx, "scheduleReviewMap data not init when get ContentsFromScheduleReview")
		return ErrorPreambleDataNotInitialized
	}

	for _, item := range scheduleReviewMap {
		if item.LiveLessonPlan == nil {
			continue
		}
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
		return err
	}

	result := make(map[string]*entity.ContentInfoInternal, len(contents))
	for _, item := range contents {
		result[item.ID] = item
	}

	at.contentMapFromScheduleReview = result

	return nil
}

func (at *AssessmentDetailInit) initOutcomesFromSchedule() error {
	ctx := at.ctx
	op := at.op

	outcomeIDMap, err := GetScheduleModel().GetLearningOutcomeIDs(ctx, op, []string{at.assessment.ScheduleID})
	if err != nil {
		return err
	}

	outcomeIDs, ok := outcomeIDMap[at.assessment.ScheduleID]
	if !ok || len(outcomeIDs) <= 0 {
		return nil
	}

	outcomes, err := GetOutcomeModel().GetByIDs(ctx, op, dbo.MustGetDB(ctx), outcomeIDs)
	if err != nil {
		return err
	}

	at.outcomesFromSchedule = outcomes

	return nil
}

// level 3
func (at *AssessmentDetailInit) initOutcomeMapFromContent() error {
	ctx := at.ctx
	op := at.op

	contents := at.contentsFromSchedule
	if contents == nil {
		log.Error(ctx, "contentsFromSchedule data not init when get OutcomeMapFromContent")
		return ErrorPreambleDataNotInitialized
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
		return err
	}

	result := make(map[string]*entity.Outcome, len(outcomes))

	for _, item := range outcomes {
		result[item.ID] = item
	}

	at.outcomeMapFromContent = result

	return nil
}

func (at *AssessmentDetailInit) initRoomData() error {
	ctx := at.ctx

	roomData := at.liveRoom
	if roomData == nil {
		log.Error(ctx, "liveRoom data not init when process room data")
		return ErrorPreambleDataNotInitialized
	}

	roomUserScoreMap, contentTree, err := GetAssessmentExternalService().StudentScores(ctx, roomData.ScoresByUser)
	if err != nil {
		return err
	}

	at.roomUserScoreMap = roomUserScoreMap
	at.roomContentTree = contentTree

	return nil
}

func (at *AssessmentDetailInit) initCommentResultMap() error {
	ctx := at.ctx
	//op := at.op

	result := make(map[string]string)

	reviewerFeedbackMap := at.reviewerFeedbackMap
	if reviewerFeedbackMap == nil {
		log.Error(ctx, "reviewerFeedbackMap data not init when get CommentResultMap")
		return ErrorPreambleDataNotInitialized
	}

	if len(reviewerFeedbackMap) > 0 {
		for _, item := range reviewerFeedbackMap {
			result[item.AssessmentUserID] = item.ReviewerComment
		}
	} else {
		studentRoomInfo := at.liveRoom
		if studentRoomInfo == nil {
			log.Error(ctx, "liveRoom data not init when get CommentResultMap")
			return ErrorPreambleDataNotInitialized
		}

		for _, item := range studentRoomInfo.TeacherCommentsByStudent {
			if item.User == nil {
				log.Warn(ctx, "get user comment error,user is empty", log.Any("studentRoomInfo", studentRoomInfo))
				continue
			}

			if len(item.TeacherComments) <= 0 {
				continue
			}

			latestComment := item.TeacherComments[len(item.TeacherComments)-1]

			result[item.User.UserID] = latestComment.Comment
		}
	}

	at.commentResultMap = result

	return nil
}

func (at *AssessmentDetailInit) initOutcomeFromAssessment() error {
	ctx := at.ctx

	assessmentUsers := at.assessmentUsers
	if assessmentUsers == nil {
		log.Error(ctx, "assessmentUsers data not init when get OutcomeFromAssessment")
		return ErrorPreambleDataNotInitialized
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
	err := assessmentV2.GetAssessmentUserOutcomeDA().Query(ctx, userOutcomeCond, &userOutcomes)
	if err != nil {
		log.Error(ctx, "query assessment user outcome error", log.Err(err), log.Any("userOutcomeCond", userOutcomeCond))
		return err
	}

	result := make(map[string]*v2.AssessmentUserOutcome, len(userOutcomes))
	for _, item := range userOutcomes {
		key := at.GetKey([]string{
			item.AssessmentUserID,
			item.AssessmentContentID,
			item.OutcomeID,
		})
		result[key] = item
	}

	at.outcomeMapFromAssessment = result

	return nil
}

func (at *AssessmentDetailInit) initContentMapFromLiveRoom() error {
	ctx := at.ctx
	//op := adc.op

	roomContents := at.roomContentTree

	result := make(map[string]*RoomContentTree, len(roomContents))
	if ok, _ := at.IsNeedConvertLatestContent(); ok {
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
			return err
		}

		for _, item := range roomContents {
			result[latestContentIDMap[item.ContentID]] = item
		}
	} else {
		for _, item := range roomContents {
			result[item.ContentID] = item
		}
	}

	at.contentMapFromLiveRoom = result

	return nil
}

func (at *AssessmentDetailInit) summaryRoomScores(userScoreMap map[string][]*RoomUserScore, contentsReply []*v2.AssessmentContentReply) (map[string]float64, map[string]float64) {
	contentSummaryTotalScoreMap := make(map[string]float64)
	contentMap := make(map[string]*v2.AssessmentContentReply)
	for _, content := range contentsReply {
		if content.FileType == v2.AssessmentFileTypeHasChildContainer {
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
			key := at.GetKey([]string{
				userID,
				resultItem.ContentUniqueID,
			})
			roomUserResultMap[key] = resultItem

			if contentItem, ok := contentMap[resultItem.ContentUniqueID]; ok {
				if contentItem.FileType == v2.AssessmentFileTypeHasChildContainer {
					continue
				}

				contentID := contentItem.ContentID
				if contentItem.ContentType == v2.AssessmentContentTypeUnknown {
					contentID = contentItem.ParentID
				}

				key2 := at.GetKey([]string{
					userID,
					contentID,
				})
				roomUserSummaryScoreMap[key2] = roomUserSummaryScoreMap[key2] + resultItem.Score
			}
		}
	}

	log.Debug(at.ctx, "summary score info", log.Any("contentSummaryTotalScoreMap", contentSummaryTotalScoreMap), log.Any("roomUserSummaryScoreMap", roomUserSummaryScoreMap))
	return contentSummaryTotalScoreMap, roomUserSummaryScoreMap
}

func (at *AssessmentDetailInit) AsyncInitData() error {
	ctx := at.ctx

	for _, levels := range at.detailInitMap {
		g := new(errgroup.Group)
		for _, levelItem := range levels {
			g.Go(levelItem)
		}

		if err := g.Wait(); err != nil {
			log.Error(ctx, "get assessment level info error",
				log.Err(err))
			return err
		}
	}

	return nil
}

func (at *AssessmentDetailInit) MatchTeacherIDs(processorMap map[v2.AssessmentType]IAssessmentProcessor) []string {
	assessmentUsers := at.assessmentUsers

	result := make([]string, 0)
	processor := processorMap[at.assessment.AssessmentType]

	for _, assUserItem := range assessmentUsers {
		if id, ok := processor.ProcessTeacherID(assUserItem); ok {
			result = append(result, id)
		}
	}

	return result
}

func (at *AssessmentDetailInit) MatchOutcomes() map[string]*v2.AssessmentOutcomeReply {
	contents := at.contentsFromSchedule
	outcomeMap := at.outcomeMapFromContent

	result := make(map[string]*v2.AssessmentOutcomeReply, len(outcomeMap))

	for _, item := range outcomeMap {
		result[item.ID] = &v2.AssessmentOutcomeReply{
			OutcomeID:          item.ID,
			OutcomeName:        item.Name,
			AssignedTo:         nil,
			Assumed:            item.Assumed,
			AssignedToLessPlan: false,
			AssignedToMaterial: false,
			ScoreThreshold:     item.ScoreThreshold,
		}
	}

	for _, materialItem := range contents {
		for _, outcomeID := range materialItem.OutcomeIDs {
			if outcomeItem, ok := result[outcomeID]; ok {
				if materialItem.ContentType == v2.AssessmentContentTypeLessonPlan {
					outcomeItem.AssignedToLessPlan = true
				}
				if materialItem.ContentType == v2.AssessmentContentTypeLessonMaterial {
					outcomeItem.AssignedToMaterial = true
				}
			}
		}
	}

	for _, outcomeItem := range result {
		if outcomeItem.AssignedToLessPlan {
			outcomeItem.AssignedTo = append(outcomeItem.AssignedTo, v2.AssessmentOutcomeAssignTypeLessonPlan)
		}
		if outcomeItem.AssignedToMaterial {
			outcomeItem.AssignedTo = append(outcomeItem.AssignedTo, v2.AssessmentOutcomeAssignTypeLessonMaterial)
		}
	}

	return result
}

func ConvertAssessmentDetailReply2(ctx context.Context, op *entity.Operator, assessment *v2.Assessment) (*v2.AssessmentDetailReply, error) {
	detailInit, err := NewAssessmentDetailInit(ctx, op, assessment)
	if err != nil {
		return nil, err
	}

	if err = detailInit.AsyncInitData(); err != nil {
		return nil, err
	}

	processorMap := make(map[v2.AssessmentType]IAssessmentProcessor)
	processorMap[v2.AssessmentTypeOnlineClass] = NewOnlineClassAssessment()
	processorMap[v2.AssessmentTypeOfflineClass] = NewOnlineClassAssessment()
	processorMap[v2.AssessmentTypeOnlineStudy] = NewOnlineStudyAssessment()
	processorMap[v2.AssessmentTypeReviewStudy] = NewReviewStudyAssessment()
	processorMap[v2.AssessmentTypeOfflineStudy] = NewOfflineStudyAssessment()

	schedule := detailInit.schedule

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

	result.Class = detailInit.class

	result.TeacherIDs = detailInit.MatchTeacherIDs(processorMap)

	result.Program = detailInit.program
	result.Subjects = detailInit.subjects
	result.ScheduleTitle = schedule.Title
	result.ScheduleDueAt = schedule.DueAt
	var remainingTime int64
	if schedule.DueAt != 0 {
		remainingTime = schedule.DueAt - time.Now().Unix()
	} else {
		remainingTime = time.Unix(detailInit.assessment.CreateAt, 0).Add(config.Get().Assessment.DefaultRemainingTime).Unix() - time.Now().Unix()
	}
	if remainingTime < 0 {
		remainingTime = 0
	}
	result.RemainingTime = remainingTime

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
