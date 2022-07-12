package model

import (
	"context"
	"database/sql"
	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/da"
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

type AssessmentInit struct {
	ctx           context.Context
	op            *entity.Operator
	assessment    *v2.Assessment
	detailInitMap map[assessmentInitLevel][]assessmentInitFunc

	//lai *AssessmentListInit

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
	scheduleStuReviewMap map[string]*entity.ScheduleReview         // key:StudentID

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

	contentMapFromScheduleReview map[string]*entity.ContentInfoInternal // key:contentID

	roomUserScoreMap map[string][]*RoomUserScore // key: userID
	roomContentTree  []*RoomContentTree

	assessmentUserIDTypeMap map[string]*v2.AssessmentUser // key: userID+userType
}

func NewAssessmentInit(ctx context.Context, op *entity.Operator, assessment *v2.Assessment) (*AssessmentInit, error) {
	if assessment == nil {
		return nil, constant.ErrRecordNotFound
	}

	at := &AssessmentInit{
		ctx:           ctx,
		op:            op,
		assessment:    assessment,
		detailInitMap: make(map[assessmentInitLevel][]assessmentInitFunc),
	}

	return at, nil
}
func NewAssessmentInitWithDetail(ctx context.Context, op *entity.Operator, assessment *v2.Assessment) (*AssessmentInit, error) {
	if assessment == nil {
		return nil, constant.ErrRecordNotFound
	}

	//lai, err := NewAssessmentListInit(ctx, op, []*v2.Assessment{assessment})
	//if err != nil {
	//	return nil, err
	//}

	at := &AssessmentInit{
		ctx:        ctx,
		op:         op,
		assessment: assessment,
		//lai:           lai,
		detailInitMap: make(map[assessmentInitLevel][]assessmentInitFunc),
	}

	at.ConfigAssessmentDetailInitMap()

	return at, nil
}

func (at *AssessmentInit) ConfigAssessmentDetailInitMap() {
	data := make(map[assessmentInitLevel][]assessmentInitFunc)
	data[assessmentInitLevel1] = append(data[assessmentInitLevel1],
		at.initSchedule,
		at.initScheduleRelation,
		at.initScheduleReviewMap,
		at.initAssessmentUsers,
		at.initLiveRoom,
		at.initOutcomesFromSchedule)

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
		at.initRoomData,
		at.initOutcomeFromAssessment)

	data[assessmentInitLevel3] = append(data[assessmentInitLevel3],
		at.initOutcomeMapFromContent,
		at.initCommentResultMap,
		at.initContentMapFromLiveRoom)

	at.detailInitMap = data
}

// Level 1
func (at *AssessmentInit) initSchedule() error {
	if at.schedule != nil {
		return nil
	}

	ctx := at.ctx
	schedules, err := GetScheduleModel().QueryUnsafe(ctx, &entity.ScheduleQueryCondition{
		IDs: entity.NullStrings{
			Strings: []string{at.assessment.ScheduleID},
			Valid:   true,
		},
	})
	if err != nil {
		log.Error(ctx, "get schedule info error", log.Err(err), log.Any("assessment", at.assessment))
		return err
	}

	if len(schedules) <= 0 {
		return constant.ErrRecordNotFound
	}

	at.schedule = schedules[0]

	return nil
}

func (at *AssessmentInit) initScheduleRelation() error {
	if at.scheduleRelations != nil {
		return nil
	}

	ctx := at.ctx
	op := at.op

	scheduleRelations, err := GetScheduleRelationModel().Query(ctx, op, &da.ScheduleRelationCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: []string{at.assessment.ScheduleID},
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

	at.scheduleRelations = scheduleRelations

	return nil
}

func (at *AssessmentInit) initScheduleReviewMap() error {
	if at.scheduleStuReviewMap != nil {
		return nil
	}

	at.scheduleStuReviewMap = make(map[string]*entity.ScheduleReview)
	if at.assessment.AssessmentType != v2.AssessmentTypeReviewStudy {
		return nil
	}

	ctx := at.ctx
	op := at.op

	scheduleID := at.assessment.ScheduleID

	scheduleReviewMap, err := GetScheduleModel().GetSuccessScheduleReview(ctx, op, []string{scheduleID})
	if err != nil {
		return err
	}

	studentReviews, ok := scheduleReviewMap[scheduleID]
	if !ok {
		log.Warn(ctx, "schedule review data invalid", log.Any("assessment", at.assessment))
		return nil
	}

	for _, reviewItem := range studentReviews {
		if reviewItem.LiveLessonPlan == nil {
			log.Warn(ctx, "student review content is empty", log.Any("studentReviewContent", reviewItem))
			continue
		}
		at.scheduleStuReviewMap[reviewItem.StudentID] = reviewItem
	}

	return nil
}

func (at *AssessmentInit) initAssessmentUsers() error {
	if at.assessmentUsers != nil {
		return nil
	}

	ctx := at.ctx

	var result []*v2.AssessmentUser
	err := assessmentV2.GetAssessmentUserDA().Query(ctx, &assessmentV2.AssessmentUserCondition{
		AssessmentIDs: entity.NullStrings{
			Strings: []string{at.assessment.ID},
			Valid:   true,
		},
	}, &result)
	if err != nil {
		return err
	}

	at.assessmentUsers = result

	return nil
}

func GetAssessmentKey(value []string) string {
	return strings.Join(value, "_")
}

func (at *AssessmentInit) initLiveRoom() error {
	if at.liveRoom != nil {
		return nil
	}

	ctx := at.ctx
	op := at.op

	if at.assessment.AssessmentType != v2.AssessmentTypeOnlineClass &&
		at.assessment.AssessmentType != v2.AssessmentTypeOnlineStudy &&
		at.assessment.AssessmentType != v2.AssessmentTypeReviewStudy {
		at.liveRoom = &external.RoomInfo{}
		return nil
	}

	roomDataMap, err := external.GetAssessmentServiceProvider().Get(ctx, op,
		[]string{at.assessment.ScheduleID},
		external.WithAssessmentGetCompletionPercentages(),
		external.WithAssessmentGetScore(),
		external.WithAssessmentGetTeacherComment())
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

func (at *AssessmentInit) initOutcomesFromSchedule() error {
	if at.outcomesFromSchedule != nil {
		return nil
	}
	ctx := at.ctx
	op := at.op

	if at.assessment.AssessmentType != v2.AssessmentTypeOfflineStudy {
		at.outcomesFromSchedule = make([]*entity.Outcome, 0)
		return nil
	}

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

// Level 2
func (at *AssessmentInit) initProgram() error {
	if at.program != nil {
		return nil
	}

	if err := at.initSchedule(); err != nil {
		return err
	}
	ctx := at.ctx
	op := at.op

	programID := at.schedule.ProgramID

	programMap, err := external.GetProgramServiceProvider().BatchGetNameMap(ctx, op, []string{programID})
	if err != nil {
		return err
	}

	at.program = &entity.IDName{
		ID:   programID,
		Name: programMap[programID],
	}

	return nil
}

func (at *AssessmentInit) initSubject() error {
	if at.subjects != nil {
		return nil
	}

	ctx := at.ctx
	op := at.op

	if err := at.initScheduleRelation(); err != nil {
		return err
	}

	relations := at.scheduleRelations
	subjectIDs := make([]string, 0)
	deDupMap := make(map[string]struct{})

	for _, srItem := range relations {
		if srItem.RelationType != entity.ScheduleRelationTypeSubject {
			continue
		}

		if _, ok := deDupMap[srItem.RelationID]; ok || srItem.RelationID == "" {
			continue
		}

		subjectIDs = append(subjectIDs, srItem.RelationID)
		deDupMap[srItem.RelationID] = struct{}{}
	}

	subjectMap, err := external.GetSubjectServiceProvider().BatchGetNameMap(ctx, op, subjectIDs)
	if err != nil {
		return err
	}

	at.subjects = make([]*entity.IDName, 0, len(subjectMap))
	for id, name := range subjectMap {
		at.subjects = append(at.subjects, &entity.IDName{
			ID:   id,
			Name: name,
		})
	}

	return nil
}

func (at *AssessmentInit) initClass() error {
	if at.class != nil {
		return nil
	}

	ctx := at.ctx
	op := at.op

	if err := at.initScheduleRelation(); err != nil {
		return err
	}

	var classID string
	for _, srItem := range at.scheduleRelations {
		if srItem.RelationType == entity.ScheduleRelationTypeClassRosterClass {
			classID = srItem.RelationID
			break
		}
	}

	at.class = &entity.IDName{}
	if classID == "" {
		return nil
	}

	classMap, err := external.GetClassServiceProvider().BatchGetNameMap(ctx, op, []string{classID})
	if err != nil {
		return err
	}

	at.class.ID = classID
	at.class.Name = classMap[classID]

	return nil
}

func (at *AssessmentInit) IsNeedConvertLatestContent() (bool, error) {
	ctx := at.ctx

	if at.schedule == nil {
		if err := at.initSchedule(); err != nil {
			return false, err
		}
	}

	if (at.assessment.MigrateFlag == constant.AssessmentHistoryFlag &&
		at.assessment.Status != v2.AssessmentStatusNotStarted) || !at.schedule.IsLockedLessonPlan() {
		log.Debug(ctx, "assessment belongs to the migration or schedule can not locked lessPlan", log.Any("assessment", at.assessment))
		return true, nil
	}

	return false, nil
}

func (at *AssessmentInit) initAssessmentUserWithIDTypeMap() error {
	if at.assessmentUserIDTypeMap != nil {
		return nil
	}

	if err := at.initAssessmentUsers(); err != nil {
		return err
	}
	assessmentUsers := at.assessmentUsers

	result := make(map[string]*v2.AssessmentUser, len(assessmentUsers))
	for _, item := range assessmentUsers {
		result[GetAssessmentKey([]string{item.UserID, item.UserType.String()})] = item
	}

	at.assessmentUserIDTypeMap = result

	return nil
}

func (at *AssessmentInit) getLatestContentsFromSchedule(schedule *entity.Schedule) ([]*v2.AssessmentContentView, error) {
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
func (at *AssessmentInit) convertContentOutcome(contents []*v2.AssessmentContentView) error {
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
func (at *AssessmentInit) getLockedContentsFromSchedule(schedule *entity.Schedule) ([]*v2.AssessmentContentView, error) {
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
func (at *AssessmentInit) getLockedContentBySchedule(schedule *entity.Schedule) ([]*v2.AssessmentContentView, error) {
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
func (at *AssessmentInit) initContentsFromSchedule() error {
	if at.contentsFromSchedule != nil {
		return nil
	}
	at.contentsFromSchedule = make([]*v2.AssessmentContentView, 0)

	if at.assessment.AssessmentType != v2.AssessmentTypeOnlineClass &&
		at.assessment.AssessmentType != v2.AssessmentTypeOfflineClass &&
		at.assessment.AssessmentType != v2.AssessmentTypeOnlineStudy {
		return nil
	}

	if err := at.initSchedule(); err != nil {
		return err
	}

	if ok, _ := at.IsNeedConvertLatestContent(); ok {
		result, err := at.getLatestContentsFromSchedule(at.schedule)
		if err != nil {
			return err
		}
		at.contentsFromSchedule = result
	} else {
		result, err := at.getLockedContentsFromSchedule(at.schedule)
		if err != nil {
			return err
		}
		at.contentsFromSchedule = result
	}
	return nil
}

func (at *AssessmentInit) initAssessmentContentMap() error {
	if at.contentMapFromAssessment != nil {
		return nil
	}

	ctx := at.ctx

	if at.assessment.AssessmentType != v2.AssessmentTypeOnlineClass &&
		at.assessment.AssessmentType != v2.AssessmentTypeOfflineClass &&
		at.assessment.AssessmentType != v2.AssessmentTypeOnlineStudy {

		at.contentMapFromAssessment = make(map[string]*v2.AssessmentContent)
		return nil
	}

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

func (at *AssessmentInit) initReviewerFeedbackMap() error {
	if at.reviewerFeedbackMap != nil {
		return nil
	}

	ctx := at.ctx

	if err := at.initAssessmentUsers(); err != nil {
		return err
	}

	assessmentUserIDs := make([]string, len(at.assessmentUsers))
	for i, item := range at.assessmentUsers {
		assessmentUserIDs[i] = item.ID
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

	at.reviewerFeedbackMap = make(map[string]*v2.AssessmentReviewerFeedback, len(feedbacks))

	for _, item := range feedbacks {
		at.reviewerFeedbackMap[item.AssessmentUserID] = item
	}

	return nil
}

func (at *AssessmentInit) initContentsFromScheduleReview() error {
	if at.contentMapFromScheduleReview != nil {
		return nil
	}

	ctx := at.ctx
	//op := at.op

	if at.assessment.AssessmentType != v2.AssessmentTypeReviewStudy {
		at.contentMapFromScheduleReview = make(map[string]*entity.ContentInfoInternal)
		return nil
	}

	if err := at.initScheduleReviewMap(); err != nil {
		return err
	}
	contentIDs := make([]string, 0)
	dedupContentID := make(map[string]struct{})

	for _, item := range at.scheduleStuReviewMap {
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

func (at *AssessmentInit) initRoomData() error {
	if at.roomUserScoreMap != nil {
		return nil
	}

	ctx := at.ctx

	if err := at.initLiveRoom(); err != nil {
		return err
	}

	roomData := at.liveRoom

	if len(roomData.ScoresByUser) <= 0 {
		at.roomUserScoreMap = make(map[string][]*RoomUserScore)
		at.roomContentTree = make([]*RoomContentTree, 0)
		return nil
	}

	roomUserScoreMap, contentTree, err := GetAssessmentExternalService().StudentScores(ctx, roomData.ScoresByUser)
	if err != nil {
		return err
	}

	at.roomUserScoreMap = roomUserScoreMap
	at.roomContentTree = contentTree

	return nil
}

func (at *AssessmentInit) initOutcomeFromAssessment() error {
	if at.outcomeMapFromAssessment != nil {
		return nil
	}
	ctx := at.ctx

	if at.assessment.AssessmentType == v2.AssessmentTypeReviewStudy {
		at.outcomeMapFromAssessment = make(map[string]*v2.AssessmentUserOutcome)
		return nil
	}

	if err := at.initAssessmentUsers(); err != nil {
		return err
	}
	assessmentUsers := at.assessmentUsers
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
		key := GetAssessmentKey([]string{
			item.AssessmentUserID,
			item.AssessmentContentID,
			item.OutcomeID,
		})
		result[key] = item
	}

	at.outcomeMapFromAssessment = result

	return nil
}

// level 3
func (at *AssessmentInit) initOutcomeMapFromContent() error {
	if at.outcomeMapFromContent != nil {
		return nil
	}
	ctx := at.ctx
	op := at.op

	if at.assessment.AssessmentType != v2.AssessmentTypeOnlineClass &&
		at.assessment.AssessmentType != v2.AssessmentTypeOfflineClass &&
		at.assessment.AssessmentType != v2.AssessmentTypeOnlineStudy {

		at.outcomeMapFromContent = make(map[string]*entity.Outcome)
		return nil
	}

	if err := at.initContentsFromSchedule(); err != nil {
		return err
	}
	contents := at.contentsFromSchedule

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

func (at *AssessmentInit) initCommentResultMap() error {
	if at.commentResultMap != nil {
		return nil
	}

	ctx := at.ctx
	//op := at.op

	result := make(map[string]string)

	if err := at.initReviewerFeedbackMap(); err != nil {
		return err
	}
	reviewerFeedbackMap := at.reviewerFeedbackMap

	if len(reviewerFeedbackMap) > 0 {
		for _, item := range reviewerFeedbackMap {
			result[item.AssessmentUserID] = item.ReviewerComment
		}
	} else {
		if err := at.initLiveRoom(); err != nil {
			return err
		}

		for _, item := range at.liveRoom.TeacherCommentsByStudent {
			if item.User == nil {
				log.Warn(ctx, "get user comment error,user is empty", log.Any("studentRoomInfo", at.liveRoom))
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

func (at *AssessmentInit) initContentMapFromLiveRoom() error {
	if at.contentMapFromLiveRoom != nil {
		return nil
	}

	ctx := at.ctx
	//op := adc.op

	if err := at.initRoomData(); err != nil {
		return err
	}

	roomContents := at.roomContentTree
	if len(roomContents) <= 0 {
		at.contentMapFromLiveRoom = make(map[string]*RoomContentTree)
		return nil
	}

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

func (at *AssessmentInit) summaryRoomScores(userScoreMap map[string][]*RoomUserScore, contentsReply []*v2.AssessmentContentReply) (map[string]float64, map[string]float64) {
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
			key := GetAssessmentKey([]string{
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

				key2 := GetAssessmentKey([]string{
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

func (at *AssessmentInit) AsyncInitData() error {
	ctx := at.ctx

	for _, levelFuncs := range at.detailInitMap {
		g := new(errgroup.Group)
		for _, levelFunc := range levelFuncs {
			g.Go(levelFunc)
		}

		if err := g.Wait(); err != nil {
			log.Error(ctx, "get assessment level info error",
				log.Err(err))
			return err
		}
	}

	return nil
}

func (at *AssessmentInit) MatchTeacherIDs() []string {
	assessmentUsers := at.assessmentUsers

	result := make([]string, 0)
	processor := AssessmentProcessorMap[at.assessment.AssessmentType]

	for _, assUserItem := range assessmentUsers {
		if id, ok := processor.ProcessTeacherID(assUserItem); ok {
			result = append(result, id)
		}
	}

	return result
}

func (at *AssessmentInit) MatchOutcomes() map[string]*v2.AssessmentOutcomeReply {
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

func ConvertAssessmentDetailReply(ctx context.Context, op *entity.Operator, assessment *v2.Assessment) (*v2.AssessmentDetailReply, error) {
	detailInit, err := NewAssessmentInitWithDetail(ctx, op, assessment)
	if err != nil {
		return nil, err
	}

	if err = detailInit.AsyncInitData(); err != nil {
		return nil, err
	}

	schedule := detailInit.schedule

	outcomeMap := detailInit.MatchOutcomes()

	contents := make([]*v2.AssessmentContentReply, 0)
	if processor, ok := AssessmentProcessorMap[assessment.AssessmentType]; ok {
		contents, err = processor.ProcessContents(ctx, detailInit)
		if err != nil {
			return nil, err
		}
	}
	students := make([]*v2.AssessmentStudentReply, 0)
	if processor, ok := AssessmentProcessorMap[assessment.AssessmentType]; ok {
		students, err = processor.ProcessStudents(ctx, detailInit, contents)
		if err != nil {
			return nil, err
		}
	}

	diffContentStudents := make([]*v2.AssessmentDiffContentStudentsReply, 0)
	if processor, ok := AssessmentProcessorMap[assessment.AssessmentType]; ok {
		diffContentStudents = processor.ProcessDiffContents(ctx, detailInit)
	}

	var isAnyOneAttempted bool
	for _, item := range detailInit.assessmentUsers {
		if item.StatusBySystem != v2.AssessmentUserSystemStatusNotStarted {
			isAnyOneAttempted = true
		}
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

	result.Class = detailInit.class

	result.TeacherIDs = detailInit.MatchTeacherIDs()

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

	if detailInit.liveRoom != nil {
		result.CompleteRate = detailInit.liveRoom.CompletionPercentage
	}
	if assessment.AssessmentType == v2.AssessmentTypeOfflineStudy {
		offlineStudy := &OfflineStudyAssessment{}
		result.CompleteRate = offlineStudy.ProcessCompleteRate(ctx, detailInit.assessmentUsers, detailInit.reviewerFeedbackMap)
	}

	result.IsAnyOneAttempted = isAnyOneAttempted || len(detailInit.roomUserScoreMap) > 0
	result.Description = schedule.Description

	for _, item := range outcomeMap {
		result.Outcomes = append(result.Outcomes, item)
	}

	result.DiffContentStudents = diffContentStudents

	return result, nil
}
