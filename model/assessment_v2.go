package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da/assessmentV2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/mutex"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	assessmentModelV2Instance     IAssessmentModelV2
	assessmentModelV2InstanceOnce = sync.Once{}

	ErrAssessmentNotAllowDelete = errors.New("assessment has been processed and cannot be deleted")
	//ErrAssessmentHasCompleted   = errors.New("assessment has completed")
)

type assessmentModelV2 struct {
	AmsServices external.AmsServices
}

type AssessmentConfigFunc func() error

type IAssessmentModelV2 interface {
	ScheduleEndClassCallback(ctx context.Context, operator *entity.Operator, args *v2.ScheduleEndClassCallBackReq) error
	AddWhenCreateSchedules(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, req *v2.AssessmentAddWhenCreateSchedulesReq) error
	Draft(ctx context.Context, op *entity.Operator, req *v2.AssessmentUpdateReq) error
	Complete(ctx context.Context, op *entity.Operator, req *v2.AssessmentUpdateReq) error

	Page(ctx context.Context, op *entity.Operator, input *v2.AssessmentQueryReq) (*v2.AssessmentPageReply, error)
	GetByID(ctx context.Context, op *entity.Operator, id string) (*v2.AssessmentDetailReply, error)
	// home page
	StatisticsCount(ctx context.Context, op *entity.Operator, req *v2.StatisticsCountReq) (*v2.AssessmentsSummary, error)
	QueryTeacherFeedback(ctx context.Context, op *entity.Operator, condition *v2.StudentQueryAssessmentConditions) (int64, []*v2.StudentAssessment, error)
	PageForHomePage(ctx context.Context, op *entity.Operator, req *v2.AssessmentQueryReq) (*v2.ListAssessmentsResultForHomePage, error)

	AnyoneAttemptedByScheduleIDs(ctx context.Context, op *entity.Operator, scheduleIDs []string) (map[string]*v2.AssessmentAnyoneAttemptedReply, error)
	QueryInternal(ctx context.Context, op *entity.Operator, condition *assessmentV2.AssessmentCondition) ([]*v2.Assessment, error)

	LockAssessmentContentAndOutcome(ctx context.Context, op *entity.Operator, schedule *entity.Schedule) error

	InternalDeleteByScheduleIDsTx(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, scheduleIDs []string) error
}

func GetAssessmentModelV2() IAssessmentModelV2 {
	assessmentModelV2InstanceOnce.Do(func() {
		assessmentModelV2Instance = &assessmentModelV2{
			AmsServices: external.GetAmsServices(),
		}
	})
	return assessmentModelV2Instance
}

func (a *assessmentModelV2) QueryInternal(ctx context.Context, op *entity.Operator, condition *assessmentV2.AssessmentCondition) ([]*v2.Assessment, error) {
	var assessments []*v2.Assessment
	err := assessmentV2.GetAssessmentDA().Query(ctx, condition, &assessments)
	if err != nil {
		log.Error(ctx, "query assessment error", log.Err(err), log.Any("condition", condition), log.Any("op", op))
		return nil, err
	}

	return assessments, nil
}

func (a *assessmentModelV2) AnyoneAttemptedByScheduleIDs(ctx context.Context, op *entity.Operator, scheduleIDs []string) (map[string]*v2.AssessmentAnyoneAttemptedReply, error) {
	var assessments []*v2.Assessment
	condition := &assessmentV2.AssessmentCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
	}
	err := assessmentV2.GetAssessmentDA().Query(ctx, condition, &assessments)
	if err != nil {
		log.Error(ctx, "query assessment error", log.Err(err), log.Any("condition", condition), log.Any("op", op))
		return nil, err
	}
	assessmentIDs := make([]string, len(assessments))
	for i, item := range assessments {
		assessmentIDs[i] = item.ID
	}

	assessmentUserCond := &assessmentV2.AssessmentUserCondition{
		AssessmentIDs: entity.NullStrings{
			Strings: assessmentIDs,
			Valid:   true,
		},
		StatusBySystem: sql.NullString{
			String: v2.AssessmentUserStatusParticipate.String(),
			Valid:  true,
		},
	}
	var assessmentUsers []*v2.AssessmentUser
	err = assessmentV2.GetAssessmentUserDA().Query(ctx, assessmentUserCond, &assessmentUsers)
	if err != nil {
		log.Error(ctx, "query assessment user error", log.Err(err), log.Any("assessmentUserCond", assessmentUserCond), log.Any("op", op))
		return nil, err
	}
	assessmentUserMap := make(map[string]bool)
	for _, item := range assessmentUsers {
		if _, ok := assessmentUserMap[item.AssessmentID]; !ok {
			assessmentUserMap[item.AssessmentID] = true
		}
	}

	result := make(map[string]*v2.AssessmentAnyoneAttemptedReply, len(assessments))
	for _, item := range assessments {
		resultItem := &v2.AssessmentAnyoneAttemptedReply{
			IsAnyoneAttempted: assessmentUserMap[item.ID],
			AssessmentStatus:  item.Status,
		}

		result[item.ScheduleID] = resultItem
	}

	return result, nil
}

func (a *assessmentModelV2) QueryTeacherFeedback(ctx context.Context, op *entity.Operator, condition *v2.StudentQueryAssessmentConditions) (int64, []*v2.StudentAssessment, error) {
	assessmentType, err := condition.ClassType.ToAssessmentType(ctx)
	if err != nil {
		return 0, nil, constant.ErrInvalidArgs
	}

	if condition.Page < 0 || condition.PageSize < 0 {
		log.Warn(ctx, "condition page or pageSize invalid", log.Any("condition", condition))
		return 0, nil, constant.ErrInvalidArgs
	}

	if assessmentType == v2.AssessmentTypeOfflineStudy {
		total, userResults, err := assessmentV2.GetAssessmentUserResultDA().GetAssessmentUserResultDBView(ctx, &assessmentV2.AssessmentUserResultDBViewCondition{
			OrgID: sql.NullString{
				String: condition.OrgID,
				Valid:  true,
			},
			UserIDs: entity.NullStrings{
				Strings: []string{condition.StudentID},
				Valid:   true,
			},
			Pager: dbo.Pager{
				Page:     condition.Page,
				PageSize: condition.PageSize,
			},
		})
		if err != nil {
			return 0, nil, err
		}

		scheduleIDs := make([]string, 0)
		feedbackIDs := make([]string, 0)
		teacherIDs := make([]string, 0)

		dedupMap := make(map[string]struct{})
		for _, item := range userResults {
			if _, ok := dedupMap[item.ScheduleID]; !ok {
				scheduleIDs = append(scheduleIDs, item.ScheduleID)
			}
			if _, ok := dedupMap[item.StudentFeedbackID]; !ok {
				feedbackIDs = append(feedbackIDs, item.StudentFeedbackID)
			}
			if _, ok := dedupMap[item.ReviewerID]; !ok && item.ReviewerID != "" {
				teacherIDs = append(teacherIDs, item.ReviewerID)
			}

			dedupMap[item.ScheduleID] = struct{}{}
			dedupMap[item.StudentFeedbackID] = struct{}{}
			dedupMap[item.ReviewerID] = struct{}{}
		}

		scheduleMap, err := a.querySchedulesMap(ctx, scheduleIDs)
		if err != nil {
			return 0, nil, err
		}
		feedbackMap, err := a.queryFeedbackInfo(ctx, op, feedbackIDs)
		if err != nil {
			return 0, nil, err
		}
		teacherMap, err := external.GetUserServiceProvider().BatchGetMap(ctx, op, teacherIDs)
		if err != nil {
			return 0, nil, err
		}

		result := make([]*v2.StudentAssessment, len(userResults))
		for i, item := range userResults {
			resultItem := &v2.StudentAssessment{
				ID:                  item.ID,
				Title:               item.Title,
				Score:               int(item.AssessScore),
				Status:              item.Status.Compliant(ctx),
				CreateAt:            item.CreateAt,
				UpdateAt:            item.UpdateAt,
				CompleteAt:          item.CompleteAt,
				TeacherComments:     nil,
				Schedule:            nil,
				FeedbackAttachments: nil,
			}
			teacherComment := &v2.StudentAssessmentTeacher{
				Teacher: &v2.StudentAssessmentTeacherInfo{
					ID:         item.ReviewerID,
					GivenName:  "",
					FamilyName: "",
					Avatar:     "",
				},
				Comment: item.ReviewerComment,
			}

			if teacherInfo, ok := teacherMap[item.ReviewerID]; ok {
				teacherComment.Teacher.GivenName = teacherInfo.GivenName
				teacherComment.Teacher.FamilyName = teacherInfo.FamilyName
				teacherComment.Teacher.Avatar = teacherInfo.Avatar
			}
			resultItem.TeacherComments = append(resultItem.TeacherComments, teacherComment)

			if scheduleInfo, ok := scheduleMap[item.ScheduleID]; ok {
				scheduleAttachment := new(v2.StudentAssessmentAttachment)
				err := json.Unmarshal([]byte(scheduleInfo.Attachment), scheduleAttachment)
				if err != nil {
					log.Error(ctx, "Unmarshal schedule attachment failed",
						log.Err(err),
						log.Any("scheduleInfo", scheduleInfo),
					)
					return 0, nil, err
				}
				resultItem.Schedule = &v2.StudentAssessmentSchedule{
					ID:         scheduleInfo.ID,
					Title:      scheduleInfo.Title,
					Type:       string(scheduleInfo.ClassType),
					Attachment: scheduleAttachment,
				}
			}
			if feedbackAttachments, ok := feedbackMap[item.StudentFeedbackID]; ok {
				for _, attachment := range feedbackAttachments {
					resultItem.FeedbackAttachments = append(resultItem.FeedbackAttachments, v2.StudentAssessmentAttachment{
						ID:   attachment.AttachmentID,
						Name: attachment.AttachmentName,
					})
				}
			}

			result[i] = resultItem
		}

		return total, result, nil
	}

	return 0, nil, nil
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
			condition.TeacherIDs.Strings = append(condition.TeacherIDs.Strings, permission.MyPermission.UserID)
		}

		condition.TeacherIDs.Valid = true
	}

	log.Debug(ctx, "permission info", log.Any("permission", permission), log.Any("condition", condition))

	return condition, nil
}

func (a *assessmentModelV2) GetAssessmentDetailConfig(adc *AssessmentDetailComponent, assessmentType v2.AssessmentType) []AssessmentConfigFunc {
	switch assessmentType {
	case v2.AssessmentTypeOnlineClass:
		return []AssessmentConfigFunc{
			adc.apc.MatchSchedule,
			adc.apc.MatchLessPlan,
			adc.apc.MatchProgram,
			adc.apc.MatchSubject,
			adc.apc.MatchTeacher,
			adc.apc.MatchClass,

			adc.MatchOutcome,
			adc.MatchContentsContainsRoomInfo,
			adc.MatchStudentContainsRoomInfo,
		}
	case v2.AssessmentTypeOfflineClass:
		return []AssessmentConfigFunc{
			adc.apc.MatchSchedule,
			adc.apc.MatchLessPlan,
			adc.apc.MatchProgram,
			adc.apc.MatchSubject,
			adc.apc.MatchTeacher,
			adc.apc.MatchClass,

			adc.MatchOutcome,
			adc.MatchContentsNotContainsRoomInfo,
			adc.MatchStudentNotContainsRoomInfo,
		}
	case v2.AssessmentTypeOnlineStudy:
		return []AssessmentConfigFunc{
			adc.apc.MatchSchedule,
			adc.apc.MatchLessPlan,
			adc.apc.MatchTeacher,
			adc.apc.MatchClass,
			adc.apc.MatchRemainingTime,
			adc.apc.MatchCompleteRate,

			adc.MatchOutcome,
			adc.MatchContentsContainsRoomInfo,
			adc.MatchStudentContainsRoomInfo,
		}
	}

	return nil
}
func (a *assessmentModelV2) GetByID(ctx context.Context, op *entity.Operator, id string) (*v2.AssessmentDetailReply, error) {
	// assessment data
	assessment := new(v2.Assessment)
	err := assessmentV2.GetAssessmentDA().Get(ctx, id, assessment)
	if err != nil {
		log.Error(ctx, "get assessment by id from db error", log.Err(err), log.String("assessmentId", id))
		return nil, err
	}

	if assessment.AssessmentType == v2.AssessmentTypeOfflineStudy {
		log.Warn(ctx, "assessment type is not support offline study", log.Err(err), log.Any("assessment", assessment))
		return nil, nil
	}
	assessmentComponent := NewAssessmentDetailComponent(ctx, op, assessment)
	result, err := assessmentComponent.ConvertDetailReply(a.GetAssessmentDetailConfig(assessmentComponent, assessment.AssessmentType))
	if err != nil {
		log.Error(ctx, "ConvertPageReply error", log.Err(err))
		return nil, err
	}

	return result, nil
}

func (a *assessmentModelV2) GetAssessmentPageConfig(ac *AssessmentPageComponent, assessmentType v2.AssessmentType) []AssessmentConfigFunc {
	switch assessmentType {
	case v2.AssessmentTypeOnlineClass, v2.AssessmentTypeOfflineClass:
		return []AssessmentConfigFunc{
			ac.MatchSchedule,
			ac.MatchLessPlan,
			ac.MatchProgram,
			ac.MatchSubject,
			ac.MatchTeacher,
		}
	case v2.AssessmentTypeOnlineStudy:
		return []AssessmentConfigFunc{
			ac.MatchSchedule,
			ac.MatchLessPlan,
			ac.MatchTeacher,
			ac.MatchClass,
			ac.MatchCompleteRate,
			ac.MatchRemainingTime,
		}
	default:
		return []AssessmentConfigFunc{
			ac.MatchTeacher,
		}
	}

	return nil
}
func (a *assessmentModelV2) Page(ctx context.Context, op *entity.Operator, req *v2.AssessmentQueryReq) (*v2.AssessmentPageReply, error) {
	condition, err := a.getConditionByPermission(ctx, op)
	if err != nil {
		return nil, err
	}

	condition.AssessmentType = sql.NullString{
		String: req.AssessmentType.String(),
		Valid:  true,
	}

	condition.Status.Strings = strings.Split(req.Status, ",")
	condition.Status.Valid = len(condition.Status.Strings) > 0
	condition.OrderBy = assessmentV2.NewAssessmentOrderBy(req.OrderBy)
	condition.Pager = dbo.Pager{
		Page:     req.PageIndex,
		PageSize: req.PageSize,
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
	}

	var assessments []*v2.Assessment
	total, err := assessmentV2.GetAssessmentDA().Page(ctx, condition, &assessments)
	if err != nil {
		log.Error(ctx, "page assessment error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}

	assessmentComponent := NewPageComponent(ctx, op, assessments)
	result, err := assessmentComponent.ConvertPageReply(a.GetAssessmentPageConfig(assessmentComponent, req.AssessmentType))
	if err != nil {
		log.Error(ctx, "ConvertPageReply error", log.Err(err))
		return nil, err
	}

	return &v2.AssessmentPageReply{
		Total:       total,
		Assessments: result,
	}, nil
}

func (a *assessmentModelV2) InternalDeleteByScheduleIDsTx(ctx context.Context, op *entity.Operator, tx *dbo.DBContext, scheduleIDs []string) error {
	var assessments []*v2.Assessment
	err := assessmentV2.GetAssessmentDA().Query(ctx, &assessmentV2.AssessmentCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
	}, &assessments)
	if err != nil {
		log.Error(ctx, "get assessment by schedule ids error", log.Err(err), log.Strings("ScheduleIDs", scheduleIDs))
		return err
	}

	assessmentIDs := make([]string, len(assessments))
	for i, item := range assessments {
		//if item.Status != v2.AssessmentStatusNotStarted {
		//	return ErrAssessmentNotAllowDelete
		//}
		assessmentIDs[i] = item.ID
	}

	// delete assessment
	err = assessmentV2.GetAssessmentDA().DeleteByScheduleIDsTx(ctx, tx, scheduleIDs)
	if err != nil {
		log.Error(ctx, "del assessment by schedule ids error", log.Err(err), log.Strings("ScheduleIDs", scheduleIDs))
		return err
	}
	// delete assessment user
	err = assessmentV2.GetAssessmentUserDA().DeleteByAssessmentIDsTx(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "del assessment user by assessment ids error", log.Err(err), log.Strings("assessmentIDs", assessmentIDs))
		return err
	}
	// delete assessment content
	err = assessmentV2.GetAssessmentContentDA().DeleteByAssessmentIDsTx(ctx, tx, assessmentIDs)
	if err != nil {
		log.Error(ctx, "del assessment content by assessment ids error", log.Err(err), log.Strings("assessmentIDs", assessmentIDs))
		return err
	}

	return err
}

func (a *assessmentModelV2) Draft(ctx context.Context, op *entity.Operator, req *v2.AssessmentUpdateReq) error {
	return a.Update(ctx, op, v2.AssessmentStatusInDraft, req)
}

func (a *assessmentModelV2) Complete(ctx context.Context, op *entity.Operator, req *v2.AssessmentUpdateReq) error {
	return a.Update(ctx, op, v2.AssessmentStatusComplete, req)
}

func (a *assessmentModelV2) Update(ctx context.Context, op *entity.Operator, status v2.AssessmentStatus, req *v2.AssessmentUpdateReq) error {
	if len(req.Students) <= 0 {
		log.Warn(ctx, "students is empty", log.Any("req", req))
		return constant.ErrInvalidArgs
	}

	permission := new(AssessmentPermission)

	err := permission.IsAllowEdit(ctx, op)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
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

	detailComponent := NewAssessmentDetailComponent(ctx, op, waitUpdatedAssessment)

	userIDAndUserTypeMap, err := detailComponent.GetAssessmentUserWithUserIDAndUserTypeMap()
	if err != nil {
		return err
	}

	waitUpdatedUsers := make([]*v2.AssessmentUser, 0)
	for _, item := range req.Students {
		existItem, ok := userIDAndUserTypeMap[detailComponent.getKey([]string{item.StudentID, v2.AssessmentUserTypeStudent.String()})]
		if !ok {
			log.Warn(ctx, "student not exist", log.Any("userIDAndUserTypeMap", userIDAndUserTypeMap), log.Any("reqItem", item))
			return constant.ErrInvalidArgs
		}
		if !item.Status.Valid() {
			log.Warn(ctx, "student status invalid", log.Any("userIDAndUserTypeMap", userIDAndUserTypeMap), log.Any("reqItem", item))
			return constant.ErrInvalidArgs
		}
		existItem.StatusByUser = item.Status
		waitUpdatedUsers = append(waitUpdatedUsers, existItem)
	}

	scheduleContents, err := detailComponent.GetContentsFromSchedule()
	if err != nil {
		return err
	}

	assessmentContentMap, err := detailComponent.GetAssessmentContentMap()
	if err != nil {
		return err
	}

	waitAddContentMap := make(map[string]*v2.AssessmentContent)
	for _, item := range scheduleContents {
		if _, ok := assessmentContentMap[item.ID]; !ok {
			waitAddContentMap[item.ID] = &v2.AssessmentContent{
				ID:           utils.NewID(),
				AssessmentID: waitUpdatedAssessment.ID,
				ContentID:    item.ID,
				ContentType:  item.ContentType,
				Status:       v2.AssessmentContentStatusNotCovered,
				CreateAt:     now,
			}
		}
	}

	waitUpdateContents := make([]*v2.AssessmentContent, 0, len(assessmentContentMap))
	for _, item := range req.Contents {
		if contentItem, ok := assessmentContentMap[item.ContentID]; ok {
			if !item.Status.Valid() {
				log.Warn(ctx, "content status is invalid", log.Any("item", item), log.Any("req.Contents", req.Contents))
				return constant.ErrInvalidArgs
			}
			contentItem.Status = item.Status
			contentItem.ReviewerComment = item.ReviewerComment
			contentItem.UpdateAt = now
			waitUpdateContents = append(waitUpdateContents, contentItem)
		} else {
			if waitAddContentItem, ok := waitAddContentMap[item.ContentID]; ok {
				if !item.Status.Valid() {
					log.Warn(ctx, "content status is invalid", log.Any("item", item), log.Any("req.Contents", req.Contents))
					return constant.ErrInvalidArgs
				}
				waitAddContentItem.ReviewerComment = item.ReviewerComment
				waitAddContentItem.Status = item.Status
			}
		}
	}

	waitAddContents := make([]*v2.AssessmentContent, 0, len(waitAddContentMap))
	for _, item := range waitAddContentMap {
		waitAddContents = append(waitAddContents, item)
	}
	allAssessmentContents := append(waitUpdateContents, waitAddContents...)

	// outcome
	contentIDs := make([]string, 0, len(waitUpdateContents)+len(waitAddContentMap))
	for _, item := range allAssessmentContents {
		contentIDs = append(contentIDs, item.ContentID)
	}

	contentOutcomeIDMap, err := detailComponent.getContentOutcomeIDsMap(contentIDs)
	if err != nil {
		return err
	}
	outcomeIDs := make([]string, 0)
	for _, item := range contentOutcomeIDMap {
		outcomeIDs = append(outcomeIDs, item...)
	}
	outcomeIDs = utils.SliceDeduplicationExcludeEmpty(outcomeIDs)
	outcomes, err := GetOutcomeModel().GetByIDs(ctx, op, dbo.MustGetDB(ctx), outcomeIDs)
	if err != nil {
		return err
	}
	outcomeMap := make(map[string]*entity.Outcome)
	for _, item := range outcomes {
		outcomeMap[item.ID] = item
	}

	outcomeFromAssessmentMap, err := detailComponent.GetOutcomeFromAssessment()
	if err != nil {
		return err
	}

	waitAddAssessmentOutcomeMap := make(map[string]*v2.AssessmentUserOutcome)
	for _, userItem := range userIDAndUserTypeMap {
		if userItem.UserType == v2.AssessmentUserTypeTeacher {
			continue
		}
		for _, contentItem := range allAssessmentContents {
			if outcomeIDs, ok := contentOutcomeIDMap[contentItem.ContentID]; ok {
				for _, outcomeID := range outcomeIDs {
					if outcomeItem, ok := outcomeMap[outcomeID]; ok {
						key := detailComponent.getKey([]string{userItem.ID, contentItem.ID, outcomeID})
						if _, ok := outcomeFromAssessmentMap[key]; ok {
							continue
						}
						waitAddOutcomeItem := &v2.AssessmentUserOutcome{
							ID:                  utils.NewID(),
							AssessmentUserID:    userItem.ID,
							AssessmentContentID: contentItem.ID,
							OutcomeID:           outcomeID,
							CreateAt:            now,
						}
						if outcomeItem.Assumed {
							waitAddOutcomeItem.Status = v2.AssessmentUserOutcomeStatusAchieved
						}

						waitAddAssessmentOutcomeMap[key] = waitAddOutcomeItem
					}
				}
			}
		}
	}

	// user comment,score, outcomes
	newScores := make([]*external.H5PSetScoreRequest, 0)
	newComments := make([]*external.H5PAddRoomCommentRequest, 0)

	contentReqMap := make(map[string]*v2.AssessmentUpdateContentReq)
	for _, item := range req.Contents {
		contentReqMap[item.ContentID] = item
	}

	allAssessmentContentMap := make(map[string]*v2.AssessmentContent)
	for _, item := range allAssessmentContents {
		allAssessmentContentMap[item.ContentID] = item
	}
	waitUpdateAssessmentOutcomes := make([]*v2.AssessmentUserOutcome, 0)

	for _, stuItem := range req.Students {
		if stuItem.Status == v2.AssessmentUserStatusNotParticipate {
			continue
		}
		// verify student data
		assessmentUserItem, ok := userIDAndUserTypeMap[detailComponent.getKey([]string{stuItem.StudentID, v2.AssessmentUserTypeStudent.String()})]
		if !ok {
			log.Warn(ctx, "student not exist", log.Any("userIDAndUserTypeMap", userIDAndUserTypeMap), log.Any("stuItem", stuItem))
			return constant.ErrInvalidArgs
		}

		for _, stuResult := range stuItem.Results {
			// verify student content data
			if assessmentContentItem, ok := allAssessmentContentMap[stuResult.ContentID]; ok {
				for _, outcomeItem := range stuResult.Outcomes {
					if !outcomeItem.Status.Valid() {
						log.Warn(ctx, "student outcome status invalid", log.Any("req", req), log.Any("outcomeItem", outcomeItem))
						return constant.ErrInvalidArgs
					}
					key := detailComponent.getKey([]string{assessmentUserItem.ID, assessmentContentItem.ID, outcomeItem.OutcomeID})
					if outcomeFromAssessmentItem, ok := outcomeFromAssessmentMap[key]; ok {
						outcomeFromAssessmentItem.Status = outcomeItem.Status
						outcomeFromAssessmentItem.UpdateAt = now
						waitUpdateAssessmentOutcomes = append(waitUpdateAssessmentOutcomes, outcomeFromAssessmentItem)
					} else if waitAddOutcomeItem, ok := waitAddAssessmentOutcomeMap[key]; ok {
						waitAddOutcomeItem.Status = outcomeItem.Status
					} else {
						log.Warn(ctx, "student outcome invalid", log.Any("outcomeFromAssessmentMap", outcomeFromAssessmentMap), log.Any("waitAddAssessmentOutcomeMap", waitAddAssessmentOutcomeMap), log.Any("stuItem", stuItem))
						continue
					}
				}
			}
			if contentItem, ok := contentReqMap[stuResult.ContentID]; ok {
				if contentItem.ParentID != "" {
					newScore := &external.H5PSetScoreRequest{
						RoomID:       waitUpdatedAssessment.ScheduleID,
						StudentID:    stuItem.StudentID,
						ContentID:    contentItem.ParentID,
						SubContentID: contentItem.ContentID,
						Score:        stuResult.Score,
					}
					newScores = append(newScores, newScore)
				}
			}
		}

		newComment := external.H5PAddRoomCommentRequest{
			RoomID:    waitUpdatedAssessment.ScheduleID,
			StudentID: stuItem.StudentID,
			Comment:   stuItem.ReviewerComment,
		}
		newComments = append(newComments, &newComment)
	}

	if waitUpdatedAssessment.AssessmentType == v2.AssessmentTypeOnlineClass ||
		waitUpdatedAssessment.AssessmentType == v2.AssessmentTypeOnlineStudy {
		// update student comment
		if len(newComments) > 0 {
			if _, err := external.GetH5PRoomCommentServiceProvider().BatchAdd(ctx, op, newComments); err != nil {
				log.Warn(ctx, "set student comment error", log.Err(err), log.Any("newComments", newComments))
			}
		}

		// update student score
		if len(newScores) > 0 {
			if _, err := external.GetH5PRoomScoreServiceProvider().BatchSet(ctx, op, newScores); err != nil {
				log.Warn(ctx, "set student score error", log.Err(err), log.Any("newScores", newScores))
			}
		}
	}

	waitUpdatedAssessment.UpdateAt = now
	waitUpdatedAssessment.Status = status
	if status == v2.AssessmentStatusComplete {
		waitUpdatedAssessment.CompleteAt = now
	}

	waitAddAssessmentOutcomes := make([]*v2.AssessmentUserOutcome, 0, len(waitAddAssessmentOutcomeMap))
	for _, item := range waitAddAssessmentOutcomeMap {
		waitAddAssessmentOutcomes = append(waitAddAssessmentOutcomes, item)
	}

	log.Debug(ctx, "wait update contents",
		log.Any("waitAddAssessmentOutcomeMap", waitAddAssessmentOutcomeMap),
		log.Any("waitUpdateAssessmentOutcomes", waitUpdateAssessmentOutcomes),
		log.Any("outcomeFromAssessmentMap", outcomeFromAssessmentMap))

	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		if _, err := assessmentV2.GetAssessmentDA().UpdateTx(ctx, tx, waitUpdatedAssessment); err != nil {
			return err
		}

		if len(waitUpdatedUsers) > 0 {
			log.Debug(ctx, "wait update users", log.Any("waitUpdatedUsers", waitUpdatedUsers))
			if _, err = assessmentV2.GetAssessmentUserDA().UpdateTx(ctx, tx, waitUpdatedUsers); err != nil {
				return err
			}
		}

		if len(waitAddContents) > 0 {
			log.Debug(ctx, "wait add contents", log.Any("waitAddContents", waitAddContents))
			if _, err = assessmentV2.GetAssessmentContentDA().InsertTx(ctx, tx, waitAddContents); err != nil {
				return err
			}
		}

		if len(waitUpdateContents) > 0 {
			log.Debug(ctx, "wait update contents", log.Any("waitUpdateContents", waitUpdateContents))
			if _, err = assessmentV2.GetAssessmentContentDA().UpdateTx(ctx, tx, waitUpdateContents); err != nil {
				return err
			}
		}

		if len(waitAddAssessmentOutcomes) > 0 {
			log.Debug(ctx, "wait add outcomes", log.Any("waitAddAssessmentOutcomes", waitAddAssessmentOutcomes))
			if _, err = assessmentV2.GetAssessmentUserOutcomeDA().InsertTx(ctx, tx, waitAddAssessmentOutcomes); err != nil {
				return err
			}
		}

		if len(waitUpdateAssessmentOutcomes) > 0 {
			log.Debug(ctx, "wait update outcomes", log.Any("waitUpdateAssessmentOutcomes", waitUpdateAssessmentOutcomes))
			if _, err = assessmentV2.GetAssessmentUserOutcomeDA().UpdateTx(ctx, tx, waitUpdateAssessmentOutcomes); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (a *assessmentModelV2) getAssessmentByScheduleID(ctx context.Context, scheduleID string) (*v2.Assessment, error) {
	var assessments []*v2.Assessment
	condition := &assessmentV2.AssessmentCondition{
		ScheduleID: sql.NullString{
			String: scheduleID,
			Valid:  true,
		},
	}
	err := assessmentV2.GetAssessmentDA().Query(ctx, condition, &assessments)
	if err != nil {
		log.Error(ctx, "query assessment error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	if len(assessments) <= 0 {
		log.Error(ctx, "assessment not found", log.Any("condition", condition))
		return nil, constant.ErrRecordNotFound
	}
	return assessments[0], nil
}

func (a *assessmentModelV2) ScheduleEndClassCallback(ctx context.Context, op *entity.Operator, req *v2.ScheduleEndClassCallBackReq) error {
	locker, err := mutex.NewLock(ctx, da.RedisKeyPrefixScheduleID, req.ScheduleID)
	if err != nil {
		log.Error(ctx, "ScheduleEndClassCallback: lock fail",
			log.Err(err),
			log.Any("req", req),
		)
		return err
	}
	locker.Lock()
	defer locker.Unlock()

	req.AttendanceIDs = utils.SliceDeduplicationExcludeEmpty(req.AttendanceIDs)
	if req.ScheduleID == "" || len(req.AttendanceIDs) <= 0 {
		log.Warn(ctx, "request data is invalid", log.Any("req", req))
		return constant.ErrInvalidArgs
	}

	assessment, err := a.getAssessmentByScheduleID(ctx, req.ScheduleID)
	if err != nil {
		return err
	}

	if assessment.Status == v2.AssessmentStatusComplete {
		log.Warn(ctx, "assessment is completed", log.Any("assessment", assessment), log.Any("req", req))
		return nil
	}

	if assessment.AssessmentType != v2.AssessmentTypeOnlineClass &&
		assessment.AssessmentType != v2.AssessmentTypeOfflineClass &&
		assessment.AssessmentType != v2.AssessmentTypeOnlineStudy {
		log.Warn(ctx, "not support this assessment type", log.Any("assessment", assessment), log.Any("req", req))
		return constant.ErrInvalidArgs
	}

	err = a.endClassCallbackUpdateAssessment(ctx, op, req, assessment)
	if err != nil {
		return err
	}

	return nil
}

func (a *assessmentModelV2) AddWhenCreateSchedules(ctx context.Context, tx *dbo.DBContext, op *entity.Operator, req *v2.AssessmentAddWhenCreateSchedulesReq) error {
	if !req.Valid(ctx) {
		log.Warn(ctx, "req is invalid", log.Any("req", req), log.Any("op", op))
		return constant.ErrInvalidArgs
	}

	now := time.Now().Unix()
	assessments := make([]*v2.Assessment, len(req.RepeatScheduleIDs))

	// title
	title, err := req.AssessmentType.Title(ctx, v2.GenerateAssessmentTitleInput{
		ClassName:    req.ClassRosterClassName,
		ScheduleName: req.ScheduleTitle,
	})
	if err != nil {
		return err
	}

	users := make([]*v2.AssessmentUser, 0, len(req.RepeatScheduleIDs)*len(req.Users))

	for i, scheduleID := range req.RepeatScheduleIDs {
		// assessment
		assessmentItem := &v2.Assessment{
			ID:             utils.NewID(),
			OrgID:          op.OrgID,
			ScheduleID:     scheduleID,
			AssessmentType: req.AssessmentType,
			Title:          title,
			Status:         v2.AssessmentStatusNotStarted,
			CreateAt:       now,
			MigrateFlag:    constant.AssessmentCurrentFlag,
		}
		if req.AssessmentType == v2.AssessmentTypeOfflineStudy {
			assessmentItem.Status = v2.AssessmentStatusNotApplicable
		}

		assessments[i] = assessmentItem

		// assessment user
		for _, userItem := range req.Users {
			attendance := &v2.AssessmentUser{
				ID:             utils.NewID(),
				AssessmentID:   assessmentItem.ID,
				UserID:         userItem.UserID,
				UserType:       userItem.UserType,
				StatusBySystem: v2.AssessmentUserStatusNotParticipate,
				StatusByUser:   v2.AssessmentUserStatusParticipate,
				CreateAt:       now,
			}

			if req.AssessmentType == v2.AssessmentTypeOnlineClass {
				attendance.StatusByUser = v2.AssessmentUserStatusNotParticipate
			}

			users = append(users, attendance)
		}
	}

	_, err = assessmentV2.GetAssessmentDA().InsertInBatchesTx(ctx, tx, assessments, constant.AssessmentBatchPageSize)
	if err != nil {
		return err
	}

	_, err = assessmentV2.GetAssessmentUserDA().InsertInBatchesTx(ctx, tx, users, constant.AssessmentBatchPageSize)
	if err != nil {
		return err
	}

	return nil
}

func (a *assessmentModelV2) LockAssessmentContentAndOutcome(ctx context.Context, op *entity.Operator, schedule *entity.Schedule) error {
	if !schedule.IsLockedLessonPlan() {
		log.Warn(ctx, "no one attempted,don't need to lock", log.Any("schedule", schedule))
		return nil
	}
	assessment, err := a.getAssessmentByScheduleID(ctx, schedule.ID)
	if err != nil {
		return err
	}

	now := time.Now().Unix()

	detailComponent := NewAssessmentDetailComponent(ctx, op, assessment)
	contentsFromSchedule, err := detailComponent.GetScheduleLockedContents(schedule)
	if err != nil {
		return err
	}

	assessmentUserMap, err := detailComponent.apc.GetAssessmentUserMap()
	if err != nil {
		return err
	}
	assessmentUsers, ok := assessmentUserMap[assessment.ID]
	if !ok {
		log.Error(ctx, "can not found assessment users", log.String("assessmentID", assessment.ID), log.Any("assessmentUserMap", assessmentUserMap))
		return constant.ErrRecordNotFound
	}

	// contents
	waitAddContents := make([]*v2.AssessmentContent, 0)
	assessmentContentMap := make(map[string]*v2.AssessmentContent)
	for _, item := range contentsFromSchedule {
		if _, ok := assessmentContentMap[item.ID]; !ok {
			contentNewItem := &v2.AssessmentContent{
				ID:           utils.NewID(),
				AssessmentID: assessment.ID,
				ContentID:    item.ID,
				ContentType:  item.ContentType,
				Status:       v2.AssessmentContentStatusCovered,
				CreateAt:     now,
			}
			waitAddContents = append(waitAddContents, contentNewItem)
			assessmentContentMap[item.ID] = contentNewItem
		}
	}

	// outcomes
	outcomeIDs := make([]string, 0)
	for _, item := range contentsFromSchedule {
		outcomeIDs = append(outcomeIDs, item.OutcomeIDs...)
	}

	outcomeIDs = utils.SliceDeduplicationExcludeEmpty(outcomeIDs)
	outcomes, err := GetOutcomeModel().GetByIDs(ctx, op, dbo.MustGetDB(ctx), outcomeIDs)
	if err != nil {
		return err
	}

	outcomeMap := make(map[string]*entity.Outcome)
	for _, item := range outcomes {
		outcomeMap[item.ID] = item
	}

	waitAddUserOutcomes := make([]*v2.AssessmentUserOutcome, 0)
	assessmentUserIDs := make([]string, 0, len(assessmentUsers))
	for _, userItem := range assessmentUsers {
		assessmentUserIDs = append(assessmentUserIDs, userItem.ID)

		for _, contentItem := range contentsFromSchedule {
			if assessmentContent, ok := assessmentContentMap[contentItem.ID]; ok {
				for _, outcomeID := range contentItem.OutcomeIDs {
					if outcomeItem, ok := outcomeMap[outcomeID]; ok {
						userOutcomeItem := &v2.AssessmentUserOutcome{
							ID:                  utils.NewID(),
							AssessmentUserID:    userItem.ID,
							AssessmentContentID: assessmentContent.ID,
							OutcomeID:           outcomeID,
							Status:              v2.AssessmentUserOutcomeStatusUnknown,
							CreateAt:            now,
							UpdateAt:            0,
							DeleteAt:            0,
						}
						if outcomeItem.Assumed {
							userOutcomeItem.Status = v2.AssessmentUserOutcomeStatusAchieved
						}

						waitAddUserOutcomes = append(waitAddUserOutcomes, userOutcomeItem)
					}
				}
			}
		}
	}

	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err := assessmentV2.GetAssessmentContentDA().DeleteByAssessmentIDsTx(ctx, tx, []string{assessment.ID})
		if err != nil {
			return err
		}

		err = assessmentV2.GetAssessmentUserOutcomeDA().DeleteByAssessmentUserIDsTx(ctx, tx, assessmentUserIDs)
		if err != nil {
			return err
		}

		_, err = assessmentV2.GetAssessmentContentDA().InsertTx(ctx, tx, waitAddContents)
		if err != nil {
			return err
		}

		if len(waitAddUserOutcomes) > 0 {
			_, err = assessmentV2.GetAssessmentUserOutcomeDA().InsertTx(ctx, tx, waitAddUserOutcomes)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Error(ctx, "lock assessment contents error", log.Err(err), log.Any("waitAddContents", waitAddContents), log.Any("waitAddUserOutcomes", waitAddUserOutcomes))
		return err
	}
	return nil
}

func (a *assessmentModelV2) endClassCallbackUpdateAssessment(ctx context.Context, op *entity.Operator, req *v2.ScheduleEndClassCallBackReq, assessment *v2.Assessment) error {
	attendanceReqMap := make(map[string]struct{})
	for _, item := range req.AttendanceIDs {
		attendanceReqMap[item] = struct{}{}
	}

	now := time.Now().Unix()

	attendanceCondition := &assessmentV2.AssessmentUserCondition{
		AssessmentID: sql.NullString{
			String: assessment.ID,
			Valid:  true,
		},
		UserIDs: entity.NullStrings{
			Strings: req.AttendanceIDs,
			Valid:   true,
		},
		StatusBySystem: sql.NullString{
			String: v2.AssessmentUserStatusNotParticipate.String(),
			Valid:  true,
		},
	}

	if assessment.Status == v2.AssessmentStatusNotStarted {
		// update assessment
		if assessment.AssessmentType == v2.AssessmentTypeOfflineClass ||
			assessment.AssessmentType == v2.AssessmentTypeOnlineClass {
			// update assessment title

			titleSplit := strings.SplitN(assessment.Title, "-", 2)
			if len(titleSplit) == 2 {
				var timeStr string
				if req.ClassEndAt > 0 {
					timeStr = time.Unix(req.ClassEndAt, 0).Format("20060102")
					assessment.Title = fmt.Sprintf("%s-%s", timeStr, titleSplit[1])
				}
			}
		}

		assessment.Status = v2.AssessmentStatusStarted
		assessment.UpdateAt = now
		assessment.ClassLength = req.ClassLength
		assessment.ClassEndAt = req.ClassEndAt

		return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
			_, err := assessmentV2.GetAssessmentDA().UpdateTx(ctx, tx, assessment)
			if err != nil {
				return err
			}

			if assessment.AssessmentType == v2.AssessmentTypeOnlineClass ||
				assessment.AssessmentType == v2.AssessmentTypeOnlineStudy {
				err := assessmentV2.GetAssessmentUserDA().UpdateStatusTx(ctx, dbo.MustGetDB(ctx), attendanceCondition, v2.AssessmentUserStatusParticipate)
				if err != nil {
					return err
				}
			}

			return nil
		})
	} else if assessment.Status == v2.AssessmentStatusStarted {
		err := assessmentV2.GetAssessmentUserDA().UpdateStatusTx(ctx, dbo.MustGetDB(ctx), attendanceCondition, v2.AssessmentUserStatusParticipate)
		if err != nil {
			return err
		}
	} else {
		err := assessmentV2.GetAssessmentUserDA().UpdateSystemStatusTx(ctx, dbo.MustGetDB(ctx), attendanceCondition, v2.AssessmentUserStatusParticipate)
		if err != nil {
			return err
		}
	}

	return nil
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

	assessmentComponent := NewPageComponent(ctx, op, assessments)
	pageResult, err := assessmentComponent.ConvertPageReply(a.GetAssessmentPageConfig(assessmentComponent, req.AssessmentType))
	if err != nil {
		log.Error(ctx, "ConvertPageReply error", log.Err(err))
		return nil, err
	}

	result := &v2.ListAssessmentsResultForHomePage{
		Total: total,
		Items: make([]*v2.AssessmentItemForHomePage, 0, len(pageResult)),
	}

	for _, item := range pageResult {
		result.Items = append(result.Items, &v2.AssessmentItemForHomePage{
			ID:       item.ID,
			Title:    item.Title,
			Teachers: item.Teachers,
			Status:   item.Status.ToReply(),
		})
	}

	return result, nil
}
