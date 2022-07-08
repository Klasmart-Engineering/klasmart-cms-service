package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"strings"
	"sync"

	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/da"
	"github.com/KL-Engineering/kidsloop-cms-service/da/assessmentV2"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
	"github.com/KL-Engineering/kidsloop-cms-service/external"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

var (
	assessmentModelV2Instance     IAssessmentModelV2
	assessmentModelV2InstanceOnce = sync.Once{}

	ErrAssessmentNotAllowDelete = errors.New("assessment has been processed and cannot be deleted")
	//ErrAssessmentHasCompleted   = errors.New("assessment has completed")

	AssessmentProcessorMap map[v2.AssessmentType]IAssessmentProcessor
)

type assessmentModelV2 struct {
	AmsServices external.AmsServices
}

type IAssessmentProcessor interface {
	ProcessTeacherName(assUserItem *v2.AssessmentUser, teacherMap map[string]*entity.IDName) (*entity.IDName, bool)
	ProcessTeacherID(assUserItem *v2.AssessmentUser) (string, bool)
	ProcessContents(ctx context.Context, at *AssessmentInit) ([]*v2.AssessmentContentReply, error)
	ProcessDiffContents(ctx context.Context, at *AssessmentInit) []*v2.AssessmentDiffContentStudentsReply
	ProcessStudents(context.Context, *AssessmentInit, []*v2.AssessmentContentReply) ([]*v2.AssessmentStudentReply, error)
	ProcessRemainingTime(ctx context.Context, dueAt int64, assessmentCreateAt int64) int64

	Update(ctx context.Context, op *entity.Operator, assessment *v2.Assessment, req *v2.AssessmentUpdateReq) error
}

type IAssessmentModelV2 interface {
	Update(ctx context.Context, op *entity.Operator, req *v2.AssessmentUpdateReq) error

	Page(ctx context.Context, op *entity.Operator, input *v2.AssessmentQueryReq) (*v2.AssessmentPageReply, error)
	GetByID(ctx context.Context, op *entity.Operator, id string) (*v2.AssessmentDetailReply, error)

	// home page
	StatisticsCount(ctx context.Context, op *entity.Operator, req *v2.StatisticsCountReq) (*v2.AssessmentsSummary, error)
	QueryStudentAssessment(ctx context.Context, op *entity.Operator, condition *v2.StudentQueryAssessmentConditions) (int, []*v2.StudentAssessment, error)
	PageForHomePage(ctx context.Context, op *entity.Operator, req *v2.AssessmentQueryReq) (*v2.ListAssessmentsResultForHomePage, error)
}

func GetAssessmentModelV2() IAssessmentModelV2 {
	assessmentModelV2InstanceOnce.Do(func() {
		assessmentModelV2Instance = &assessmentModelV2{
			AmsServices: external.GetAmsServices(),
		}
		AssessmentProcessorMap = make(map[v2.AssessmentType]IAssessmentProcessor)
		AssessmentProcessorMap[v2.AssessmentTypeOnlineClass] = NewOnlineClassAssessment()
		AssessmentProcessorMap[v2.AssessmentTypeOfflineClass] = NewOfflineClassAssessment()
		AssessmentProcessorMap[v2.AssessmentTypeOnlineStudy] = NewOnlineStudyAssessment()
		AssessmentProcessorMap[v2.AssessmentTypeReviewStudy] = NewReviewStudyAssessment()
		AssessmentProcessorMap[v2.AssessmentTypeOfflineStudy] = NewOfflineStudyAssessment()
	})
	return assessmentModelV2Instance
}

func (a *assessmentModelV2) Page(ctx context.Context, op *entity.Operator, req *v2.AssessmentQueryReq) (*v2.AssessmentPageReply, error) {
	condition, err := a.getConditionByPermission(ctx, op)
	if err != nil {
		return nil, err
	}

	condition.AssessmentTypes.Strings = strings.Split(req.AssessmentType.String(), ",")
	condition.AssessmentTypes.Valid = req.AssessmentType != ""

	if req.Status != "" {
		statusMap := make(map[string]struct{})
		for _, item := range condition.Status.Strings {
			statusMap[item] = struct{}{}
		}

		condition.Status.Strings = make([]string, 0)
		statusReq := strings.Split(req.Status, ",")
		for _, item := range statusReq {
			if _, ok := statusMap[item]; ok {
				condition.Status.Strings = append(condition.Status.Strings, item)
			}
		}

		condition.Status.Valid = true
	}

	condition.OrderBy = assessmentV2.NewAssessmentOrderBy(req.OrderBy)
	condition.Pager = dbo.Pager{
		Page:     req.PageIndex,
		PageSize: req.PageSize,
	}

	if req.QueryType == v2.QueryTypeTeacherID {
		// For the query key data, there is currently no check to see if there is permission to query.
		condition.TeacherIDs.Strings = []string{req.QueryKey}
		condition.TeacherIDs.Valid = true
	}
	if req.QueryType == v2.QueryTypeTeacherName {
		teachers, err := external.GetTeacherServiceProvider().Query(ctx, op, op.OrgID, req.QueryKey)
		if err != nil {
			return nil, err
		}
		condition.TeacherIDs.Valid = true
		condition.TeacherIDs.Strings = make([]string, len(teachers))
		for i, item := range teachers {
			condition.TeacherIDs.Strings[i] = item.ID
		}
		condition.TeacherIDs.Valid = true
	}

	log.Info(ctx, "condition info", log.Any("condition", condition))

	var assessments []*v2.Assessment
	total, err := assessmentV2.GetAssessmentDA().Page(ctx, condition, &assessments)
	if err != nil {
		log.Error(ctx, "page assessment error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}

	if len(assessments) <= 0 {
		return &v2.AssessmentPageReply{
			Total:       0,
			Assessments: make([]*v2.AssessmentQueryReply, 0),
		}, nil
	}

	result, err := ConvertAssessmentPageReply(ctx, op, assessments)
	if err != nil {
		return nil, err
	}

	return &v2.AssessmentPageReply{
		Total:       total,
		Assessments: result,
	}, nil
}

func (a *assessmentModelV2) GetByID(ctx context.Context, op *entity.Operator, id string) (*v2.AssessmentDetailReply, error) {
	// assessment data
	assessment := new(v2.Assessment)
	err := assessmentV2.GetAssessmentDA().Get(ctx, id, assessment)
	if err != nil {
		log.Error(ctx, "get assessment by id from db error", log.Err(err), log.String("assessmentId", id))
		return nil, err
	}

	//if assessment.AssessmentType == v2.AssessmentTypeOfflineStudy {
	//	log.Warn(ctx, "assessment type is not support offline study", log.Err(err), log.Any("assessment", assessment))
	//	return nil, nil
	//}

	result, err := ConvertAssessmentDetailReply(ctx, op, assessment)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a *assessmentModelV2) QueryStudentAssessment(ctx context.Context, op *entity.Operator, req *v2.StudentQueryAssessmentConditions) (int, []*v2.StudentAssessment, error) {
	assessmentType := v2.AssessmentType(req.ClassType)
	if req.ClassType != v2.AssessmentTypeCompliantOfflineStudy.String() {
		if !assessmentType.Valid(ctx) {
			log.Warn(ctx, "assessment class type is invalid", log.Any("req_condition", req))
			return 0, nil, constant.ErrInvalidArgs
		}
	} else {
		assessmentType = v2.AssessmentTypeOfflineStudy
	}

	// Compatible with old fields
	if req.CompleteStartAt > 0 && req.CompletedGe == 0 {
		req.CompletedGe = req.CompleteStartAt
	}
	//Compatible with old fields
	if req.CompleteEndAt > 0 && req.CompletedLe == 0 {
		req.CompletedLe = req.CompleteEndAt
	}

	condition := &assessmentV2.StudentAssessmentCondition{
		OrgID: sql.NullString{
			String: op.OrgID,
			Valid:  true,
		},
		ScheduleIDs: entity.NullStrings{
			Strings: req.ScheduleIDs,
			Valid:   len(req.ScheduleIDs) > 0,
		},
		AssessmentType: sql.NullString{
			String: assessmentType.String(),
			Valid:  true,
		},
		TeacherIDs: entity.NullStrings{
			Strings: []string{req.TeacherID},
			Valid:   req.TeacherID != "",
		},
		StudentID: sql.NullString{
			String: req.StudentID,
			Valid:  true,
		},
		Status: entity.NullStrings{
			Strings: []string{req.Status},
			Valid:   req.Status != "",
		},
		CreatedAtGe: sql.NullInt64{
			Int64: req.CreatedGe,
			Valid: req.CreatedGe > 0,
		},
		CreatedAtLe: sql.NullInt64{
			Int64: req.CreatedLe,
			Valid: req.CreatedLe > 0,
		},

		DoneAtGe: sql.NullInt64{
			Int64: req.DoneGe,
			Valid: req.DoneGe > 0,
		},
		DoneAtLe: sql.NullInt64{
			Int64: req.DoneLe,
			Valid: req.DoneLe > 0,
		},

		ResubmittedAtGe: sql.NullInt64{
			Int64: req.ResubmittedGe,
			Valid: req.ResubmittedGe > 0,
		},
		ResubmittedAtLe: sql.NullInt64{
			Int64: req.ResubmittedLe,
			Valid: req.ResubmittedLe > 0,
		},

		CompleteAtGe: sql.NullInt64{
			Int64: req.CompletedGe,
			Valid: req.CompletedGe > 0,
		},
		CompleteAtLe: sql.NullInt64{
			Int64: req.CompletedLe,
			Valid: req.CompletedLe > 0,
		},

		OrderBy: assessmentV2.StudentAssessmentOrderBy(req.OrderBy),
		Pager: dbo.Pager{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
	}

	log.Debug(ctx, "condition", log.Any("req", req), log.Any("condition", condition), log.Any("op", op))

	total, studentAssessments, err := assessmentV2.GetAssessmentDA().PageStudentAssessment(ctx, condition)
	if err != nil {
		log.Error(ctx, "query assessment error", log.Any("req", req), log.Any("condition", condition), log.Any("op", op))
		return 0, nil, err
	}

	if len(studentAssessments) <= 0 {
		return 0, make([]*v2.StudentAssessment, 0), nil
	}

	if assessmentType == v2.AssessmentTypeOfflineStudy {
		result, err := a.fillOfflineStudy(ctx, op, studentAssessments)
		if err != nil {
			return 0, nil, err
		}
		return total, result, err
	} else {
		result, err := a.fillNoneOfflineStudy(ctx, op, studentAssessments)
		if err != nil {
			return 0, nil, err
		}
		return total, result, err
	}
}

func (a *assessmentModelV2) fillNoneOfflineStudy(ctx context.Context, op *entity.Operator, stuAssessments []*v2.StudentAssessmentDBView) ([]*v2.StudentAssessment, error) {
	scheduleIDs := make([]string, len(stuAssessments))
	for i, item := range stuAssessments {
		scheduleIDs[i] = item.ScheduleID
	}
	scheduleMap, err := a.querySchedulesMap(ctx, scheduleIDs)
	if err != nil {
		return nil, err
	}

	result := make([]*v2.StudentAssessment, len(stuAssessments))

	for i, item := range stuAssessments {
		replyItem := &v2.StudentAssessment{
			ID:                  item.ID,
			Title:               item.Title,
			Type:                item.AssessmentType,
			Score:               0,
			Status:              item.StatusBySystem,
			CreateAt:            item.CreateAt,
			UpdateAt:            item.UpdateAt,
			CompleteAt:          item.CompletedAt,
			TeacherComments:     make([]*v2.StudentAssessmentTeacher, 0),
			Schedule:            new(v2.StudentAssessmentSchedule),
			FeedbackAttachments: make([]*v2.StudentAssessmentAttachment, 0),
		}
		result[i] = replyItem

		schedule, ok := scheduleMap[item.ScheduleID]
		if !ok {
			log.Warn(ctx, "not found schedule", log.Any("stuAssessmentItem", item), log.Any("scheduleMap", scheduleMap))
			continue
		}

		replyItem.Schedule = &v2.StudentAssessmentSchedule{
			ID:         schedule.ID,
			Title:      schedule.Title,
			Type:       string(schedule.ClassType),
			Attachment: new(v2.StudentScheduleAttachment),
		}
	}

	return result, nil
}

func (a *assessmentModelV2) fillNoneOfflineStudyWithRoomInfo(ctx context.Context, op *entity.Operator, assessments []*v2.Assessment, req *v2.StudentQueryAssessmentConditions) ([]*v2.StudentAssessment, error) {
	scheduleIDs := make([]string, len(assessments))
	assessmentIDs := make([]string, len(assessments))
	for i, item := range assessments {
		scheduleIDs[i] = item.ScheduleID
		assessmentIDs[i] = item.ID
	}
	scheduleMap, err := a.querySchedulesMap(ctx, scheduleIDs)
	if err != nil {
		return nil, err
	}

	userCondition := &assessmentV2.AssessmentUserCondition{
		AssessmentIDs: entity.NullStrings{
			Strings: assessmentIDs,
			Valid:   true,
		},
		UserIDs: entity.NullStrings{
			Strings: []string{req.StudentID},
			Valid:   true,
		},
		UserType: sql.NullString{
			String: v2.AssessmentUserTypeStudent.String(),
			Valid:  true,
		},
	}

	var assessmentUsers []*v2.AssessmentUser
	err = assessmentV2.GetAssessmentUserDA().Query(ctx, userCondition, &assessmentUsers)
	if err != nil {
		return nil, err
	}

	if len(assessmentUsers) <= 0 {
		return nil, constant.ErrRecordNotFound
	}

	assessmentUserIDs := make([]string, len(assessmentUsers))
	// key: assessment_id
	assessmentUserMapAssessmentIDKey := make(map[string]*v2.AssessmentUser, len(assessmentUsers))
	for i, item := range assessmentUsers {
		assessmentUserIDs[i] = item.ID
		assessmentUserMapAssessmentIDKey[item.AssessmentID] = item
	}

	commentMap, err := a.queryAssessmentComments(ctx, op, scheduleIDs, req.StudentID)
	if err != nil {
		log.Error(ctx, "queryAssessmentComments failed",
			log.Err(err),
			log.Strings("scheduleIDs", scheduleIDs),
			log.Any("req", req),
		)
		return nil, err
	}
	teacherIDs := make([]string, 0)
	for _, commentItem := range commentMap {
		for key, _ := range commentItem {
			teacherIDs = append(teacherIDs, key)
		}
	}
	teacherIDs = utils.SliceDeduplicationExcludeEmpty(teacherIDs)
	teacherMap, err := external.GetUserServiceProvider().BatchGetMap(ctx, op, teacherIDs)
	if err != nil {
		return nil, err
	}

	result := make([]*v2.StudentAssessment, len(assessments))

	for i, item := range assessments {
		replyItem := &v2.StudentAssessment{
			ID:                  item.ID,
			Title:               item.Title,
			Type:                item.AssessmentType,
			Score:               0,
			Status:              "",
			CreateAt:            item.CreateAt,
			UpdateAt:            item.UpdateAt,
			CompleteAt:          item.CompleteAt,
			TeacherComments:     make([]*v2.StudentAssessmentTeacher, 0),
			Schedule:            new(v2.StudentAssessmentSchedule),
			FeedbackAttachments: make([]*v2.StudentAssessmentAttachment, 0),
		}
		result[i] = replyItem

		schedule, ok := scheduleMap[item.ScheduleID]
		if !ok {
			continue
		}

		if assessmentUserItem, ok := assessmentUserMapAssessmentIDKey[item.ID]; ok {
			replyItem.Status = assessmentUserItem.StatusBySystem
		}

		if teacherCommentMap, ok := commentMap[item.ScheduleID]; ok {
			teacherCommentItem := &v2.StudentAssessmentTeacher{
				Teacher: new(v2.StudentAssessmentTeacherInfo),
			}
			for teacherID, comment := range teacherCommentMap {
				teacherCommentItem.Teacher.ID = teacherID
				teacherCommentItem.Comment = comment

				if teacherInfo, ok := teacherMap[teacherID]; ok {
					teacherCommentItem.Teacher.FamilyName = teacherInfo.FamilyName
					teacherCommentItem.Teacher.GivenName = teacherInfo.GivenName
					teacherCommentItem.Teacher.Avatar = teacherInfo.Avatar
				}
			}
			replyItem.TeacherComments = append(replyItem.TeacherComments, teacherCommentItem)
		}

		replyItem.Schedule = &v2.StudentAssessmentSchedule{
			ID:         schedule.ID,
			Title:      schedule.Title,
			Type:       string(schedule.ClassType),
			Attachment: new(v2.StudentScheduleAttachment),
		}
	}

	return result, nil
}

func (m *assessmentModelV2) queryAssessmentComments(ctx context.Context, operator *entity.Operator, scheduleIDs []string, studentID string) (map[string]map[string]string, error) {
	commentMap, err := getAssessmentH5P().batchGetRoomCommentObjectMap(ctx, operator, scheduleIDs)
	if err != nil {
		log.Error(ctx, "getAssessmentH5p.batchGetRoomCommentMap failed",
			log.Err(err),
			log.Strings("scheduleIDs", scheduleIDs),
		)
		return nil, err
	}
	comments := make(map[string]map[string]string)
	for i := range scheduleIDs {
		if commentMap[scheduleIDs[i]] != nil {
			studentComments := commentMap[scheduleIDs[i]][studentID]
			comments[scheduleIDs[i]] = make(map[string]string)
			for j := range studentComments {
				if studentComments[j] == nil {
					continue
				}
				comments[scheduleIDs[i]][studentComments[j].TeacherID] = studentComments[j].Comment
			}
		}
	}
	return comments, nil
}

func (a *assessmentModelV2) fillOfflineStudy(ctx context.Context, op *entity.Operator, stuAssessments []*v2.StudentAssessmentDBView) ([]*v2.StudentAssessment, error) {
	scheduleIDs := make([]string, len(stuAssessments))
	for i, item := range stuAssessments {
		scheduleIDs[i] = item.ScheduleID
	}
	scheduleMap, err := a.querySchedulesMap(ctx, scheduleIDs)
	if err != nil {
		return nil, err
	}

	assessmentUserIDs := make([]string, 0, len(stuAssessments))
	// key: id
	assessmentUserMap := make(map[string]*v2.StudentAssessmentDBView, len(stuAssessments))
	for _, item := range stuAssessments {
		assessmentUserIDs = append(assessmentUserIDs, item.ID)
		assessmentUserMap[item.ID] = item
	}
	reviewerFeedbackCond := &assessmentV2.AssessmentUserResultCondition{
		AssessmentUserIDs: entity.NullStrings{
			Strings: assessmentUserIDs,
			Valid:   true,
		}}
	var reviewerFeedbacks []*v2.AssessmentReviewerFeedback
	err = assessmentV2.GetAssessmentUserResultDA().Query(ctx, reviewerFeedbackCond, &reviewerFeedbacks)
	if err != nil {
		log.Error(ctx, "query reviewer feedback error", log.Any("reviewerFeedbackCond", reviewerFeedbackCond))
		return nil, err
	}

	// key: assessment id
	reviewerFeedbackMap := make(map[string]*v2.AssessmentReviewerFeedback)
	feedbackIDs := make([]string, 0, len(reviewerFeedbacks))
	teacherIDs := make([]string, 0)
	for _, item := range reviewerFeedbacks {
		reviewerFeedbackMap[item.AssessmentUserID] = item
		feedbackIDs = append(feedbackIDs, item.StudentFeedbackID)
		teacherIDs = append(teacherIDs, item.ReviewerID)
	}

	studentFeedbackMap, err := a.queryFeedbackInfo(ctx, op, feedbackIDs)
	if err != nil {
		return nil, err
	}

	teacherIDs = utils.SliceDeduplicationExcludeEmpty(teacherIDs)
	teacherMap, err := external.GetUserServiceProvider().BatchGetMap(ctx, op, teacherIDs)
	if err != nil {
		return nil, err
	}

	result := make([]*v2.StudentAssessment, len(stuAssessments))

	for i, item := range stuAssessments {
		replyItem := &v2.StudentAssessment{
			ID:                  item.AssessmentID,
			Title:               item.Title,
			Type:                item.AssessmentType,
			Score:               0,
			Status:              item.StatusBySystem,
			CreateAt:            item.CreateAt,
			UpdateAt:            item.UpdateAt,
			InProgressAt:        item.InProgressAt,
			DoneAt:              item.DoneAt,
			ResubmittedAt:       item.ResubmittedAt,
			CompleteAt:          item.CompletedAt,
			TeacherComments:     make([]*v2.StudentAssessmentTeacher, 0),
			Schedule:            new(v2.StudentAssessmentSchedule),
			FeedbackAttachments: make([]*v2.StudentAssessmentAttachment, 0),
		}
		result[i] = replyItem

		schedule, ok := scheduleMap[item.ScheduleID]
		if !ok {
			continue
		}

		if reviewerFeedbackItem, ok := reviewerFeedbackMap[item.ID]; ok {
			replyItem.Score = int(reviewerFeedbackItem.AssessScore)

			teacherCommentItem := &v2.StudentAssessmentTeacher{
				Teacher: new(v2.StudentAssessmentTeacherInfo),
				Comment: reviewerFeedbackItem.ReviewerComment,
			}
			teacherCommentItem.Teacher.ID = reviewerFeedbackItem.ReviewerID
			if teacherInfo, ok := teacherMap[reviewerFeedbackItem.ReviewerID]; ok {
				teacherCommentItem.Teacher.FamilyName = teacherInfo.FamilyName
				teacherCommentItem.Teacher.GivenName = teacherInfo.GivenName
				teacherCommentItem.Teacher.Avatar = teacherInfo.Avatar
			}
			replyItem.TeacherComments = append(replyItem.TeacherComments, teacherCommentItem)

			if feedbackAttachments, ok := studentFeedbackMap[reviewerFeedbackItem.StudentFeedbackID]; ok {
				for _, attachment := range feedbackAttachments {
					replyItem.FeedbackAttachments = append(replyItem.FeedbackAttachments, &v2.StudentAssessmentAttachment{
						ID:                 attachment.AttachmentID,
						Name:               attachment.AttachmentName,
						ReviewAttachmentID: attachment.ReviewAttachmentID,
					})
				}
			}
		}

		scheduleAttachment := new(v2.StudentScheduleAttachment)
		err := json.Unmarshal([]byte(schedule.Attachment), scheduleAttachment)
		if err != nil {
			log.Error(ctx, "Unmarshal schedule attachment failed",
				log.Err(err),
				log.Any("schedule", schedule),
			)
			return nil, err
		}
		replyItem.Schedule = &v2.StudentAssessmentSchedule{
			ID:         schedule.ID,
			Title:      schedule.Title,
			Type:       string(schedule.ClassType),
			Attachment: scheduleAttachment,
		}
	}

	return result, nil
}

func (a *assessmentModelV2) StatisticsCount(ctx context.Context, op *entity.Operator, req *v2.StatisticsCountReq) (*v2.AssessmentsSummary, error) {
	condition, err := a.getConditionByPermission(ctx, op)
	if err != nil {
		return nil, err
	}

	if req.Status != "" {
		condition.Status.Strings = strings.Split(req.Status, ",")
		condition.Status.Valid = len(condition.Status.Strings) > 0
	}

	var assessments []*v2.Assessment
	err = assessmentV2.GetAssessmentDA().Query(ctx, condition, &assessments)
	if err != nil {
		log.Error(ctx, "query assessment error", log.Err(err), log.Any("condition", condition), log.Any("op", op))
		return nil, err
	}
	r := new(v2.AssessmentsSummary)
	for _, a := range assessments {
		switch a.Status {
		case v2.AssessmentStatusStarted, v2.AssessmentStatusInDraft:
			r.InProgress++
		case v2.AssessmentStatusComplete:
			r.Complete++
		}

		//if a.AssessmentType == v2.AssessmentTypeOnlineStudy && a.Status == v2.AssessmentStatusNotStarted {
		//	r.InProgress++
		//}
	}

	return r, nil
}

func (a *assessmentModelV2) PageForHomePage(ctx context.Context, op *entity.Operator, req *v2.AssessmentQueryReq) (*v2.ListAssessmentsResultForHomePage, error) {
	condition, err := a.getConditionByPermission(ctx, op)
	if err != nil {
		return nil, err
	}

	condition.Status.Strings = strings.Split(req.Status, ",")
	condition.Status.Valid = len(condition.Status.Strings) > 0
	condition.OrderBy = assessmentV2.NewAssessmentOrderBy(req.OrderBy)
	condition.Pager = dbo.Pager{
		Page:     req.PageIndex,
		PageSize: req.PageSize,
	}

	var assessments []*v2.Assessment
	total, err := assessmentV2.GetAssessmentDA().Page(ctx, condition, &assessments)
	if err != nil {
		log.Error(ctx, "page assessment error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}

	if len(assessments) <= 0 {
		return &v2.ListAssessmentsResultForHomePage{
			Total: 0,
			Items: make([]*v2.AssessmentItemForHomePage, 0),
		}, nil
	}

	at, err := NewAssessmentListInit(ctx, op, assessments)
	if err != nil {
		return nil, err
	}

	err = at.initAssessmentUserMap()
	if err != nil {
		return nil, err
	}

	err = at.initTeacherMap()
	if err != nil {
		return nil, err
	}

	assessmentUserMap := at.assessmentUserMap
	teacherMap := at.teacherMap

	assTeacherMap := make(map[string][]*entity.IDName, len(assessments))
	for _, item := range assessments {
		if assUserItems, ok := assessmentUserMap[item.ID]; ok {
			for _, assUserItem := range assUserItems {
				if assUserItem.UserType != v2.AssessmentUserTypeTeacher {
					continue
				}
				if item.AssessmentType == v2.AssessmentTypeOnlineClass &&
					assUserItem.StatusByUser == v2.AssessmentUserStatusNotParticipate {
					continue
				}

				if userItem, ok := teacherMap[assUserItem.UserID]; ok && userItem != nil {
					assTeacherMap[item.ID] = append(assTeacherMap[item.ID], userItem)
				}
			}
		}
	}

	result := &v2.ListAssessmentsResultForHomePage{
		Total: total,
		Items: make([]*v2.AssessmentItemForHomePage, 0, len(assessments)),
	}

	for _, item := range assessments {
		replyItem := &v2.AssessmentItemForHomePage{
			ID:     item.ID,
			Title:  item.Title,
			Status: item.Status.ToReply(),
		}

		if teachers, ok := assTeacherMap[item.ID]; ok {
			replyItem.Teachers = teachers
		}
		result.Items = append(result.Items, replyItem)
	}

	return result, nil
}

func (a *assessmentModelV2) querySchedulesMap(ctx context.Context, scheduleIDs []string) (map[string]*entity.Schedule, error) {
	schedules, err := GetScheduleModel().QueryUnsafe(ctx, &entity.ScheduleQueryCondition{IDs: entity.NullStrings{
		Strings: scheduleIDs,
		Valid:   true,
	}})
	if err != nil {
		log.Error(ctx, "GetScheduleModel.QueryUnsafe failed",
			log.Err(err),
			log.Strings("scheduleIDs", scheduleIDs),
		)
		return nil, err
	}
	schedulesMap := make(map[string]*entity.Schedule)
	for i := range schedules {
		schedulesMap[schedules[i].ID] = schedules[i]
	}
	return schedulesMap, nil
}

func (a *assessmentModelV2) getConditionByPermission(ctx context.Context, op *entity.Operator) (*assessmentV2.AssessmentCondition, error) {
	permission := new(AssessmentPermission)

	err := permission.SearchAllPermissions(ctx, op)
	if err != nil {
		return nil, err
	}

	condition := &assessmentV2.AssessmentCondition{
		OrgID: sql.NullString{
			String: op.OrgID,
			Valid:  true,
		},
	}
	if permission.OrgPermission.Status.Valid {
		condition.Status = permission.OrgPermission.Status
	} else {
		if permission.SchoolPermission.Status.Valid {
			condition.Status = permission.SchoolPermission.Status
			condition.TeacherIDs = entity.NullStrings{
				Strings: permission.SchoolPermission.UserIDs,
				Valid:   true,
			}
		}
		if permission.MyPermission.Status.Valid {
			condition.Status.Strings = append(condition.Status.Strings, permission.MyPermission.Status.Strings...)
			condition.Status.Strings = utils.SliceDeduplicationExcludeEmpty(condition.Status.Strings)
			condition.Status.Valid = true
			condition.TeacherIDs.Strings = append(condition.TeacherIDs.Strings, permission.MyPermission.UserID)
		}

		condition.TeacherIDs.Valid = true
	}

	log.Info(ctx, "permission info", log.Any("permission", permission), log.Any("condition", condition))

	return condition, nil
}

func (a *assessmentModelV2) queryFeedbackInfo(ctx context.Context, operator *entity.Operator, feedbackIDs []string) (map[string][]*entity.FeedbackAssignmentView, error) {
	//query feedbacks
	var err error
	feedbackMap := make(map[string][]*entity.FeedbackAssignmentView)
	if len(feedbackIDs) > 0 {
		feedbackMap, err = GetFeedbackAssignmentModel().QueryMap(ctx, operator, &da.FeedbackAssignmentCondition{
			FeedBackIDs: entity.NullStrings{
				Strings: feedbackIDs,
				Valid:   true,
			},
		})
		if err != nil {
			log.Error(ctx, "GetTeacherServiceProvider.BatchGetNameMap failed",
				log.Err(err),
				log.Strings("feedbackIDs", feedbackIDs),
			)
			return nil, err
		}
	}
	return feedbackMap, nil
}

// TODO need refactor
func (a *assessmentModelV2) Update(ctx context.Context, op *entity.Operator, req *v2.AssessmentUpdateReq) error {
	if len(req.Students) <= 0 {
		log.Warn(ctx, "students is empty", log.Any("req", req))
		return constant.ErrInvalidArgs
	}

	permission := new(AssessmentPermission)

	err := permission.IsAllowEdit(ctx, op)
	if err != nil {
		return err
	}

	waitUpdatedAssessment := new(v2.Assessment)
	err = assessmentV2.GetAssessmentDA().Get(ctx, req.ID, waitUpdatedAssessment)
	if err == dbo.ErrRecordNotFound {
		return constant.ErrRecordNotFound
	}
	if err != nil {
		return err
	}
	if waitUpdatedAssessment.Status == v2.AssessmentStatusComplete {
		log.Warn(ctx, "assessment has completed", log.Any("assessment", waitUpdatedAssessment), log.Any("req", req))
		return ErrAssessmentHasCompleted
	}

	return AssessmentProcessorMap[waitUpdatedAssessment.AssessmentType].Update(ctx, op, waitUpdatedAssessment, req)
}
