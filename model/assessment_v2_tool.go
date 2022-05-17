package model

import (
	"context"
	"database/sql"
	"github.com/KL-Engineering/kidsloop-cms-service/utils/errgroup"
	"strings"
	"sync"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/da"
	"github.com/KL-Engineering/kidsloop-cms-service/da/assessmentV2"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

type AssessmentTool struct {
	ctx  context.Context
	op   *entity.Operator
	once sync.Once

	assessments []*v2.Assessment
	first       *v2.Assessment

	AssessmentBatch
	AssessmentGetOne
}

type AssessmentBatch struct {
	assessmentMap       map[string]*v2.Assessment             // assessmentID
	scheduleMap         map[string]*entity.Schedule           // scheduleID
	scheduleRelationMap map[string][]*entity.ScheduleRelation // scheduleID
	assessmentUserMap   map[string][]*v2.AssessmentUser       // assessmentID
	programMap          map[string]*entity.IDName             // programID
	subjectMap          map[string]*entity.IDName             // subjectID
	classMap            map[string]*entity.IDName             // classID
	userMap             map[string]*entity.IDName             // userID
	liveRoomMap         map[string]*external.RoomInfo         // roomID
	lessPlanMap         map[string]*v2.AssessmentContentView  // lessPlanID

	assessmentReviewerFeedbackMap map[string]*v2.AssessmentReviewerFeedback // assessmentUserID

	assessmentUsers []*v2.AssessmentUser

	scheduleStuReviewMap map[string]map[string]*entity.ScheduleReview // key:ScheduleID,StudentID
}

type AssessmentGetOne struct {
	outcomeMapFromContent      map[string]*entity.Outcome // key: outcomeID
	latestContentsFromSchedule []*v2.AssessmentContentView
	lockedContentsFromSchedule []*v2.AssessmentContentView
	contentMapFromAssessment   map[string]*v2.AssessmentContent     // key:contentID
	contentMapFromLiveRoom     map[string]*RoomContentTree          // key: contentID
	commentResultMap           map[string]string                    // userID
	outcomeMapFromAssessment   map[string]*v2.AssessmentUserOutcome // key: AssessmentUserID+AssessmentContentID+OutcomeID
	outcomesFromSchedule       []*entity.Outcome

	scheduleReviewMap            map[string]*entity.ScheduleReview      // key:StudentID
	contentMapFromScheduleReview map[string]*entity.ContentInfoInternal // key:contentID

	roomUserScoreMap map[string][]*RoomUserScore // key: userID
	roomContentTree  []*RoomContentTree

	firstAssessmentUserMap map[string]*v2.AssessmentUser // key: userID+userType
}

func NewAssessmentTool(ctx context.Context, op *entity.Operator, assessments []*v2.Assessment) (*AssessmentTool, error) {
	if len(assessments) <= 0 {
		return nil, constant.ErrRecordNotFound
	}

	return &AssessmentTool{
		ctx:         ctx,
		op:          op,
		once:        sync.Once{},
		assessments: assessments,
		first:       assessments[0],
	}, nil
}

//type ToolModel func() error

//func (at *AssessmentTool) InitAssessment(toolkit []ToolModel) error {
//	var err error
//	at.once.Do(func() {
//		for _, tool := range toolkit {
//			if err = tool(); err != nil {
//				return
//			}
//		}
//	})
//
//	return err
//}

func (at *AssessmentTool) GetAssessmentMap() (map[string]*v2.Assessment, error) {
	if at.assessmentMap == nil {
		if err := at.initAssessmentMap(); err != nil {
			return nil, err
		}
	}
	return at.assessmentMap, nil
}
func (at *AssessmentTool) initAssessmentMap() error {
	result := make(map[string]*v2.Assessment, len(at.assessments))
	for _, item := range at.assessments {
		result[item.ID] = item
	}

	at.assessmentMap = result

	return nil
}

func (at *AssessmentTool) GetScheduleMap() (map[string]*entity.Schedule, error) {
	if at.scheduleMap == nil {
		if err := at.initScheduleMap(); err != nil {
			return nil, err
		}
	}

	return at.scheduleMap, nil
}
func (at *AssessmentTool) initScheduleMap() error {
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

func (at *AssessmentTool) GetScheduleRelationMap() (map[string][]*entity.ScheduleRelation, error) {
	if at.scheduleRelationMap == nil {
		if err := at.initScheduleRelationMap(); err != nil {
			return nil, err
		}
	}
	return at.scheduleRelationMap, nil
}
func (at *AssessmentTool) initScheduleRelationMap() error {
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

func (at *AssessmentTool) GetAssessmentUsers() ([]*v2.AssessmentUser, error) {
	if at.assessmentUsers == nil {
		if err := at.initAssessmentUsers(); err != nil {
			return nil, err
		}
	}
	return at.assessmentUsers, nil
}
func (at *AssessmentTool) initAssessmentUsers() error {
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

func (at *AssessmentTool) GetAssessmentUserMap() (map[string][]*v2.AssessmentUser, error) {
	if at.assessmentUserMap == nil {
		if err := at.initAssessmentUserMap(); err != nil {
			return nil, err
		}
	}
	return at.assessmentUserMap, nil
}
func (at *AssessmentTool) initAssessmentUserMap() error {
	result := make(map[string][]*v2.AssessmentUser, len(at.assessments))

	assessmentUsers, err := at.GetAssessmentUsers()
	if err != nil {
		return err
	}

	for _, item := range assessmentUsers {
		result[item.AssessmentID] = append(result[item.AssessmentID], item)
	}

	at.assessmentUserMap = result

	return nil
}

func (at *AssessmentTool) GetProgramMap() (map[string]*entity.IDName, error) {
	if at.programMap == nil {
		err := at.AsyncInitExternalData(AssessmentExternalInclude{
			userServiceInclude: &ExternalUserServiceInclude{
				Program: true,
			},
		})
		if err != nil {
			return nil, err
		}
	}
	return at.programMap, nil
}

func (at *AssessmentTool) GetSubjectMap() (map[string]*entity.IDName, error) {
	if at.subjectMap == nil {
		err := at.AsyncInitExternalData(AssessmentExternalInclude{
			userServiceInclude: &ExternalUserServiceInclude{
				Subject: true,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	return at.subjectMap, nil
}

func (at *AssessmentTool) GetClassMap() (map[string]*entity.IDName, error) {
	if at.classMap == nil {
		err := at.AsyncInitExternalData(AssessmentExternalInclude{
			userServiceInclude: &ExternalUserServiceInclude{
				Class: true,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	return at.classMap, nil
}

func (at *AssessmentTool) GetTeacherMap() (map[string]*entity.IDName, error) {
	if at.userMap == nil {
		err := at.AsyncInitExternalData(AssessmentExternalInclude{
			userServiceInclude: &ExternalUserServiceInclude{
				Teacher: true,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	return at.userMap, nil
}

func (at *AssessmentTool) GetRoomStudentScoresAndComments() (map[string]*external.RoomInfo, error) {
	if at.liveRoomMap == nil {
		err := at.AsyncInitExternalData(AssessmentExternalInclude{
			assessmentServiceInclude: &ExternalAssessmentServiceInclude{
				StudentScore:   true,
				TeacherComment: true,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	return at.liveRoomMap, nil
}

func (at *AssessmentTool) GetLessPlanMap() (map[string]*v2.AssessmentContentView, error) {
	if at.lessPlanMap == nil {
		if err := at.initLessPlanMap(); err != nil {
			return nil, err
		}
	}

	return at.lessPlanMap, nil
}
func (at *AssessmentTool) initLessPlanMap() error {
	ctx := at.ctx

	scheduleMap, err := at.GetScheduleMap()
	if err != nil {
		return err
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
		return err
	}

	for _, latestID := range latestLassPlanIDMap {
		lockedLessPlanIDs = append(lockedLessPlanIDs, latestID)
	}
	lessPlanIDs := utils.SliceDeduplicationExcludeEmpty(lockedLessPlanIDs)
	lessPlans, err := GetContentModel().GetContentByIDListInternal(ctx, dbo.MustGetDB(ctx), lessPlanIDs)
	if err != nil {
		log.Error(ctx, "get content by ids error", log.Err(err), log.Strings("lessPlanIDs", lessPlanIDs))
		return err
	}

	result := make(map[string]*v2.AssessmentContentView, len(lessPlans))
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

func (at *AssessmentTool) GetReviewerFeedbackMap() (map[string]*v2.AssessmentReviewerFeedback, error) {
	if at.assessmentReviewerFeedbackMap == nil {
		if err := at.initReviewerFeedbackMap(); err != nil {
			return nil, err
		}
	}

	return at.assessmentReviewerFeedbackMap, nil
}
func (at *AssessmentTool) initReviewerFeedbackMap() error {
	ctx := at.ctx

	assessmentUsers, err := at.GetAssessmentUsers()
	if err != nil {
		return err
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
		return err
	}

	result := make(map[string]*v2.AssessmentReviewerFeedback)
	for _, item := range feedbacks {
		result[item.AssessmentUserID] = item
	}

	at.assessmentReviewerFeedbackMap = result

	return nil
}

func (at *AssessmentTool) BatchGetScheduleReviewMap() (map[string]map[string]*entity.ScheduleReview, error) {
	if at.scheduleStuReviewMap == nil {
		if err := at.initBatchGetScheduleReviewMap(); err != nil {
			return nil, err
		}
	}
	return at.scheduleStuReviewMap, nil
}
func (at *AssessmentTool) initBatchGetScheduleReviewMap() error {
	ctx := at.ctx
	op := at.op

	scheduleIDs := make([]string, len(at.assessments))
	for _, item := range at.assessments {
		scheduleIDs = append(scheduleIDs, item.ScheduleID)
	}

	scheduleReviewMap, err := GetScheduleModel().GetSuccessScheduleReview(ctx, op, scheduleIDs)
	if err != nil {
		return err
	}

	result := make(map[string]map[string]*entity.ScheduleReview)
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

func (at *AssessmentTool) FirstGetAssessmentUserWithUserIDAndUserTypeMap() (map[string]*v2.AssessmentUser, error) {
	if at.firstAssessmentUserMap == nil {
		if err := at.initFirstGetAssessmentUserWithUserIDAndUserTypeMap(); err != nil {
			return nil, err
		}
	}

	return at.firstAssessmentUserMap, nil
}

func (at *AssessmentTool) initFirstGetAssessmentUserWithUserIDAndUserTypeMap() error {
	ctx := at.ctx

	assessmentUserMap, err := at.GetAssessmentUserMap()
	if err != nil {
		return err
	}

	assessmentUsers, ok := assessmentUserMap[at.first.ID]
	if !ok {
		log.Error(ctx, "not found assessment users", log.Any("assessment", at.first), log.Any("assessmentUserMap", assessmentUserMap))
		return constant.ErrRecordNotFound
	}

	result := make(map[string]*v2.AssessmentUser, len(assessmentUsers))
	for _, item := range assessmentUsers {
		result[at.GetKey([]string{item.UserID, item.UserType.String()})] = item
	}

	at.firstAssessmentUserMap = result

	return nil
}

func (at *AssessmentTool) GetKey(value []string) string {
	return strings.Join(value, "_")
}

func (at *AssessmentTool) IsNeedConvertLatestContent() (bool, error) {
	ctx := at.ctx

	schedule, err := at.FirstGetSchedule()
	if err != nil {
		return false, err
	}

	if (at.first.MigrateFlag == constant.AssessmentHistoryFlag &&
		at.first.Status != v2.AssessmentStatusNotStarted) || !schedule.IsLockedLessonPlan() {
		log.Debug(ctx, "assessment belongs to the migration or schedule can not locked lessPlan", log.Any("assessment", at.first))
		return true, nil
	}

	return false, nil
}

func (at *AssessmentTool) convertContentOutcome(contents []*v2.AssessmentContentView) error {
	ctx := at.ctx
	op := at.op

	contentOutcomes, err := assessmentV2.GetAssessmentUserOutcomeDA().GetContentOutcomeByAssessmentID(ctx, at.first.ID)
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

func (at *AssessmentTool) FirstGetLatestContentsFromSchedule() ([]*v2.AssessmentContentView, error) {
	if at.latestContentsFromSchedule == nil {
		if err := at.initFirstGetLatestContentsFromSchedule(); err != nil {
			return nil, err
		}
	}

	return at.latestContentsFromSchedule, nil
}
func (at *AssessmentTool) initFirstGetLatestContentsFromSchedule() error {
	ctx := at.ctx
	op := at.op

	schedule, err := at.FirstGetSchedule()
	if err != nil {
		return err
	}

	// convert to latest lesson plan
	latestLessPlanIDMap, err := GetContentModel().GetLatestContentIDMapByIDListInternal(ctx, dbo.MustGetDB(ctx), []string{schedule.LessonPlanID})
	if err != nil {
		return err
	}
	latestLessPlanID, ok := latestLessPlanIDMap[schedule.LessonPlanID]
	if !ok {
		log.Error(ctx, "lessPlan not found", log.Any("schedule", schedule), log.Any("latestLessPlanIDMap", latestLessPlanIDMap))
		return constant.ErrRecordNotFound
	}

	// get lesson plan info
	latestLessPlans, err := GetContentModel().GetContentByIDListInternal(ctx, dbo.MustGetDB(ctx), []string{latestLessPlanID})
	if err != nil {
		return err
	}
	if len(latestLessPlans) <= 0 {
		log.Warn(ctx, "not found content info", log.String("latestLessPlanID", latestLessPlanID), log.Any("schedule", schedule))
		return constant.ErrRecordNotFound
	}

	// get material in lesson plan
	subContentsMap, err := GetContentModel().GetContentsSubContentsMapByIDListInternal(ctx, dbo.MustGetDB(ctx), []string{latestLessPlanID}, op)
	if err != nil {
		return err
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
		return err
	}

	at.latestContentsFromSchedule = result

	return nil
}

func (at *AssessmentTool) FirstGetSchedule() (*entity.Schedule, error) {
	scheduleMap, err := at.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	schedule, ok := scheduleMap[at.first.ScheduleID]
	if !ok {
		return nil, constant.ErrRecordNotFound
	}

	return schedule, nil
}

func (at *AssessmentTool) FirstGetLockedContentsFromSchedule() ([]*v2.AssessmentContentView, error) {
	if at.lockedContentsFromSchedule == nil {
		if err := at.initFirstGetLockedContentsFromSchedule(); err != nil {
			return nil, err
		}
	}

	return at.lockedContentsFromSchedule, nil
}

func (at *AssessmentTool) initFirstGetLockedContentsFromSchedule() error {
	schedule, err := at.FirstGetSchedule()
	if err != nil {
		return err
	}

	at.lockedContentsFromSchedule = make([]*v2.AssessmentContentView, 0)

	result, err := at.firstGetLockedContentBySchedule(schedule)
	if err != nil {
		return err
	}

	err = at.convertContentOutcome(result)
	if err != nil {
		return err
	}

	at.lockedContentsFromSchedule = result

	return nil
}

func (at *AssessmentTool) firstGetLockedContentBySchedule(schedule *entity.Schedule) ([]*v2.AssessmentContentView, error) {
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

func (at *AssessmentTool) FirstGetContentsFromSchedule() ([]*v2.AssessmentContentView, error) {
	var result []*v2.AssessmentContentView
	var err error
	if ok, _ := at.IsNeedConvertLatestContent(); ok {
		result, err = at.FirstGetLatestContentsFromSchedule()
	} else {
		result, err = at.FirstGetLockedContentsFromSchedule()
	}

	return result, err
}

func (at *AssessmentTool) FirstGetOutcomeMapFromContent() (map[string]*entity.Outcome, error) {
	if at.outcomeMapFromContent == nil {
		if err := at.initFirstGetOutcomeMapFromContent(); err != nil {
			return nil, err
		}
	}

	return at.outcomeMapFromContent, nil
}
func (at *AssessmentTool) initFirstGetOutcomeMapFromContent() error {
	ctx := at.ctx
	op := at.op

	contents, err := at.FirstGetContentsFromSchedule()
	if err != nil {
		return err
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

func (at *AssessmentTool) FirstGetAssessmentContentMap() (map[string]*v2.AssessmentContent, error) {
	if at.contentMapFromAssessment == nil {
		if err := at.initFirstGetAssessmentContentMap(); err != nil {
			return nil, err
		}
	}
	return at.contentMapFromAssessment, nil
}
func (at *AssessmentTool) initFirstGetAssessmentContentMap() error {
	ctx := at.ctx

	var assessmentContents []*v2.AssessmentContent
	err := assessmentV2.GetAssessmentContentDA().Query(ctx, &assessmentV2.AssessmentUserCondition{
		AssessmentID: sql.NullString{
			String: at.first.ID,
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

func (at *AssessmentTool) FirstGetContentMapFromLiveRoom() (map[string]*RoomContentTree, error) {
	if at.contentMapFromLiveRoom == nil {
		if err := at.initFirstGetContentMapFromLiveRoom(); err != nil {
			return nil, err
		}
	}

	return at.contentMapFromLiveRoom, nil
}
func (at *AssessmentTool) initFirstGetContentMapFromLiveRoom() error {
	ctx := at.ctx
	//op := adc.op

	_, roomContents, err := at.FirstGetRoomData()
	if err != nil {
		return err
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

func (at *AssessmentTool) FirstGetRoomData() (map[string][]*RoomUserScore, []*RoomContentTree, error) {
	if at.roomUserScoreMap == nil || at.roomContentTree == nil {
		if err := at.initFirstGetRoomData(); err != nil {
			return nil, nil, err
		}
	}

	return at.roomUserScoreMap, at.roomContentTree, nil
}
func (at *AssessmentTool) initFirstGetRoomData() error {
	ctx := at.ctx
	//op := adc.op

	roomDataMap, err := at.GetRoomStudentScoresAndComments()
	if err != nil {
		return err
	}

	roomData, ok := roomDataMap[at.first.ScheduleID]
	if !ok {
		log.Warn(ctx, "not found room data", log.Any("roomDataMap", roomDataMap), log.Any("assessment", at.first))

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

func (at *AssessmentTool) FirstGetCommentResultMap() (map[string]string, error) {
	if at.commentResultMap == nil {
		if err := at.initFirstGetCommentResultMap(); err != nil {
			return nil, err
		}
	}

	return at.commentResultMap, nil
}
func (at *AssessmentTool) initFirstGetCommentResultMap() error {
	ctx := at.ctx
	//op := at.op

	result := make(map[string]string)

	studentRoomInfoMap, err := at.GetRoomStudentScoresAndComments()
	if err != nil {
		return err
	}

	studentRoomInfo, ok := studentRoomInfoMap[at.first.ScheduleID]
	if !ok {
		log.Warn(ctx, "not found student room info from studentRoomInfoMap", log.Any("assessment", at.first), log.Any("studentRoomInfoMap", studentRoomInfoMap))
		return nil
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

	at.commentResultMap = result

	return nil
}

func (at *AssessmentTool) FirstGetOutcomeFromAssessment() (map[string]*v2.AssessmentUserOutcome, error) {
	if at.outcomeMapFromAssessment == nil {
		if err := at.initFirstGetOutcomeFromAssessment(); err != nil {
			return nil, err
		}
	}

	return at.outcomeMapFromAssessment, nil
}
func (at *AssessmentTool) initFirstGetOutcomeFromAssessment() error {
	ctx := at.ctx

	assessmentUserMap, err := at.GetAssessmentUserMap()
	if err != nil {
		return err
	}

	assessmentUsers, ok := assessmentUserMap[at.first.ID]
	if !ok {
		return constant.ErrRecordNotFound
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

func (at *AssessmentTool) FirstGetScheduleReviewMap() (map[string]*entity.ScheduleReview, error) {
	if at.scheduleReviewMap == nil {
		if err := at.initFirstGetScheduleReviewMap(); err != nil {
			return nil, err
		}
	}
	return at.scheduleReviewMap, nil
}
func (at *AssessmentTool) initFirstGetScheduleReviewMap() error {
	ctx := at.ctx
	//op := at.op

	scheduleReviewMap, err := at.BatchGetScheduleReviewMap()
	if err != nil {
		return err
	}

	studentReviews, ok := scheduleReviewMap[at.first.ScheduleID]
	if !ok {
		log.Warn(ctx, "schedule review info not found", log.Any("assessment", at.first))
		return constant.ErrRecordNotFound
	}

	result := make(map[string]*entity.ScheduleReview)
	for _, item := range studentReviews {
		if item.LiveLessonPlan == nil {
			continue
		}
		result[item.StudentID] = item
	}

	at.scheduleReviewMap = result

	return nil
}

func (at *AssessmentTool) FirstGetContentsFromScheduleReview() (map[string]*entity.ContentInfoInternal, error) {
	if at.contentMapFromScheduleReview == nil {
		if err := at.initFirstGetContentsFromScheduleReview(); err != nil {
			return nil, err
		}
	}

	return at.contentMapFromScheduleReview, nil
}
func (at *AssessmentTool) initFirstGetContentsFromScheduleReview() error {
	ctx := at.ctx
	//op := at.op

	contentIDs := make([]string, 0)
	dedupContentID := make(map[string]struct{})
	scheduleReviewMap, err := at.FirstGetScheduleReviewMap()
	if err != nil {
		return err
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

func (at *AssessmentTool) FirstGetOutcomesFromSchedule() ([]*entity.Outcome, error) {
	if at.outcomesFromSchedule == nil {
		if err := at.initFirstGetOutcomesFromSchedule(); err != nil {
			return nil, err
		}
	}

	return at.outcomesFromSchedule, nil
}
func (at *AssessmentTool) initFirstGetOutcomesFromSchedule() error {
	ctx := at.ctx
	op := at.op

	outcomeIDMap, err := GetScheduleModel().GetLearningOutcomeIDs(ctx, op, []string{at.first.ScheduleID})
	if err != nil {
		return err
	}

	outcomeIDs, ok := outcomeIDMap[at.first.ScheduleID]
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

func (at *AssessmentTool) summaryRoomScores(userScoreMap map[string][]*RoomUserScore, contentsReply []*v2.AssessmentContentReply) (map[string]float64, map[string]float64) {
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

type AssessmentExternalInclude struct {
	userServiceInclude       *ExternalUserServiceInclude
	assessmentServiceInclude *ExternalAssessmentServiceInclude
}

type ExternalUserServiceInclude struct {
	Program bool
	Subject bool
	Class   bool
	Teacher bool
}

type ExternalAssessmentServiceInclude struct {
	StudentScore   bool
	TeacherComment bool
}

func (at *AssessmentTool) AsyncInitExternalData(include AssessmentExternalInclude) error {
	ctx := at.ctx
	op := at.op

	g := new(errgroup.Group)

	if include.userServiceInclude != nil {
		if include.userServiceInclude.Program {
			programIDs, err := at.getProgramIDs()
			if err != nil {
				return err
			}

			g.Go(func() error {
				programs, err := external.GetProgramServiceProvider().BatchGet(ctx, op, programIDs)
				if err != nil {
					return err
				}

				result := make(map[string]*entity.IDName)
				for _, item := range programs {
					if item == nil || item.ID == "" {
						log.Warn(ctx, "program id is empty", log.Any("programs", programs))
						continue
					}
					result[item.ID] = &entity.IDName{
						ID:   item.ID,
						Name: item.Name,
					}
				}

				at.programMap = result

				return nil
			})
		}

		if include.userServiceInclude.Subject {
			subjectIDs, err := at.getSubjectIDs()
			if err != nil {
				return err
			}
			g.Go(func() error {
				subjects, err := external.GetSubjectServiceProvider().BatchGet(ctx, op, subjectIDs)
				if err != nil {
					return err
				}

				result := make(map[string]*entity.IDName)
				for _, item := range subjects {
					if item == nil || item.ID == "" {
						log.Warn(ctx, "subject is empty", log.Any("subjects", subjects))
						continue
					}
					result[item.ID] = &entity.IDName{
						ID:   item.ID,
						Name: item.Name,
					}
				}

				at.subjectMap = result

				return nil
			})
		}

		if include.userServiceInclude.Class {
			classIDs, err := at.getClassIDs()
			if err != nil {
				return err
			}
			g.Go(func() error {
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
			})
		}

		if include.userServiceInclude.Teacher {
			teacherIDs, err := at.getTeacherIDs()
			if err != nil {
				return err
			}

			g.Go(func() error {
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

				at.userMap = result

				return nil
			})
		}
	}

	if include.assessmentServiceInclude != nil {
		if include.assessmentServiceInclude.StudentScore || include.assessmentServiceInclude.TeacherComment {
			scheduleIDs, err := at.getScheduleIDs()
			if err != nil {
				return err
			}

			g.Go(func() error {
				roomDataMap, err := external.GetAssessmentServiceProvider().Get(ctx, op, scheduleIDs,
					external.WithAssessmentGetScore(include.assessmentServiceInclude.StudentScore),
					external.WithAssessmentGetTeacherComment(include.assessmentServiceInclude.TeacherComment))
				if err != nil {
					log.Warn(ctx, "external service error",
						log.Err(err), log.Strings("scheduleIDs", scheduleIDs), log.Any("op", at.op))
					at.liveRoomMap = make(map[string]*external.RoomInfo)
				} else {
					at.liveRoomMap = roomDataMap
				}

				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		log.Error(ctx, "get assessment basic info error",
			log.Err(err))
		return err
	}

	return nil
}

func (at *AssessmentTool) getProgramIDs() ([]string, error) {
	scheduleMap, err := at.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	programIDs := make([]string, 0)
	deDupMap := make(map[string]struct{})
	for _, item := range scheduleMap {
		if _, ok := deDupMap[item.ProgramID]; !ok && item.ProgramID != "" {
			programIDs = append(programIDs, item.ProgramID)
			deDupMap[item.ProgramID] = struct{}{}
		}
	}
	return programIDs, nil
}

func (at *AssessmentTool) getSubjectIDs() ([]string, error) {
	relationMap, err := at.GetScheduleRelationMap()
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

	return subjectIDs, nil
}

func (at *AssessmentTool) getClassIDs() ([]string, error) {
	relationMap, err := at.GetScheduleRelationMap()
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

	return classIDs, nil
}

func (at *AssessmentTool) getScheduleIDs() ([]string, error) {
	scheduleMap, err := at.GetScheduleMap()
	if err != nil {
		return nil, err
	}

	scheduleIDs := make([]string, 0, len(at.assessments))
	for _, item := range scheduleMap {
		scheduleIDs = append(scheduleIDs, item.ID)
	}

	return scheduleIDs, nil
}

func (at *AssessmentTool) getTeacherIDs() ([]string, error) {
	assessmentUsers, err := at.GetAssessmentUsers()
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(assessmentUsers))
	deDupMap := make(map[string]struct{})

	for _, auItem := range assessmentUsers {
		if auItem.UserType != v2.AssessmentUserTypeTeacher {
			continue
		}
		if _, ok := deDupMap[auItem.UserID]; !ok {
			deDupMap[auItem.UserID] = struct{}{}
			userIDs = append(userIDs, auItem.UserID)
		}
	}

	return userIDs, nil
}
