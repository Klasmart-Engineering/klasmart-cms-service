package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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
	Draft(ctx context.Context, op *entity.Operator, req *v2.AssessmentUpdateReq) error
	Complete(ctx context.Context, op *entity.Operator, req *v2.AssessmentUpdateReq) error

	Page(ctx context.Context, op *entity.Operator, input *v2.AssessmentQueryReq) (*v2.AssessmentPageReply, error)
	GetByID(ctx context.Context, op *entity.Operator, id string) (*v2.AssessmentDetailReply, error)

	// home page
	StatisticsCount(ctx context.Context, op *entity.Operator, req *v2.StatisticsCountReq) (*v2.AssessmentsSummary, error)
	QueryTeacherFeedback(ctx context.Context, op *entity.Operator, condition *v2.StudentQueryAssessmentConditions) (int64, []*v2.StudentAssessment, error)
	PageForHomePage(ctx context.Context, op *entity.Operator, req *v2.AssessmentQueryReq) (*v2.ListAssessmentsResultForHomePage, error)
}

func GetAssessmentModelV2() IAssessmentModelV2 {
	assessmentModelV2InstanceOnce.Do(func() {
		assessmentModelV2Instance = &assessmentModelV2{
			AmsServices: external.GetAmsServices(),
		}
	})
	return assessmentModelV2Instance
}

func (a *assessmentModelV2) Draft(ctx context.Context, op *entity.Operator, req *v2.AssessmentUpdateReq) error {
	return a.update(ctx, op, v2.AssessmentStatusInDraft, req)
}

func (a *assessmentModelV2) Complete(ctx context.Context, op *entity.Operator, req *v2.AssessmentUpdateReq) error {
	return a.update(ctx, op, v2.AssessmentStatusComplete, req)
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

	if len(assessments) <= 0 {
		return &v2.AssessmentPageReply{
			Total:       0,
			Assessments: make([]*v2.AssessmentQueryReply, 0),
		}, nil
	}

	result, err := ConvertAssessmentPageReply(ctx, op, req.AssessmentType, assessments)
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
			CompleteAtGe: sql.NullInt64{
				Int64: condition.CompleteStartAt,
				Valid: condition.CompleteStartAt > 0,
			},
			CompleteAtLe: sql.NullInt64{
				Int64: condition.CompleteEndAt,
				Valid: condition.CompleteEndAt > 0,
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

		result := make([]*v2.StudentAssessment, 0, len(userResults))
		for _, item := range userResults {
			status := item.Status.Compliant(ctx)
			resultItem := &v2.StudentAssessment{
				ID:                  item.ID,
				Title:               item.Title,
				Score:               int(item.AssessScore),
				Status:              status,
				CreateAt:            item.CreateAt,
				UpdateAt:            item.UpdateAt,
				CompleteAt:          item.CompleteAt,
				TeacherComments:     make([]*v2.StudentAssessmentTeacher, 0),
				Schedule:            nil,
				FeedbackAttachments: nil,
			}
			if item.ReviewerID != "" {
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
			}

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
						ID:                 attachment.AttachmentID,
						Name:               attachment.AttachmentName,
						ReviewAttachmentID: attachment.ReviewAttachmentID,
					})
				}
			}

			result = append(result, resultItem)
		}

		return total, result, nil
	}

	return 0, nil, nil
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

	at, err := NewAssessmentTool(ctx, op, assessments)
	if err != nil {
		return nil, err
	}

	assessmentUserMap, err := at.GetAssessmentUserMap()
	if err != nil {
		return nil, err
	}

	userMap, err := at.GetUserMap()
	if err != nil {
		return nil, err
	}

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

				if userItem, ok := userMap[assUserItem.UserID]; ok && userItem != nil {
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
			condition.TeacherIDs.Strings = append(condition.TeacherIDs.Strings, permission.MyPermission.UserID)
		}

		condition.TeacherIDs.Valid = true
	}

	log.Debug(ctx, "permission info", log.Any("permission", permission), log.Any("condition", condition))

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
func (a *assessmentModelV2) update(ctx context.Context, op *entity.Operator, status v2.AssessmentStatus, req *v2.AssessmentUpdateReq) error {
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

	at, err := NewAssessmentTool(ctx, op, []*v2.Assessment{waitUpdatedAssessment})
	if err != nil {
		return err
	}
	if waitUpdatedAssessment.AssessmentType == v2.AssessmentTypeOfflineStudy {
		match := GetAssessmentDetailMatch(waitUpdatedAssessment.AssessmentType, at)
		return match.Update(req)
	}

	userIDAndUserTypeMap, err := at.FirstGetAssessmentUserWithUserIDAndUserTypeMap()
	if err != nil {
		return err
	}

	waitUpdatedUsers := make([]*v2.AssessmentUser, 0)
	for _, item := range req.Students {
		existItem, ok := userIDAndUserTypeMap[at.GetKey([]string{item.StudentID, v2.AssessmentUserTypeStudent.String()})]
		if !ok {
			log.Warn(ctx, "student not exist", log.Any("userIDAndUserTypeMap", userIDAndUserTypeMap), log.Any("reqItem", item))
			return constant.ErrInvalidArgs
		}
		if !item.Status.Valid() {
			log.Warn(ctx, "student status invalid", log.Any("userIDAndUserTypeMap", userIDAndUserTypeMap), log.Any("reqItem", item))
			return constant.ErrInvalidArgs
		}
		existItem.StatusByUser = item.Status

		if req.Action == v2.AssessmentActionComplete {
			if existItem.StatusBySystem == v2.AssessmentUserSystemStatusDone || existItem.StatusBySystem == v2.AssessmentUserSystemStatusResubmitted {
				existItem.StatusBySystem = v2.AssessmentUserSystemStatusCompleted
			}
		}
		existItem.UpdateAt = now
		waitUpdatedUsers = append(waitUpdatedUsers, existItem)
	}

	roomDataMap, err := at.GetRoomStudentScoresAndComments()
	if err != nil {
		return err
	}
	roomData, hasScore := roomDataMap[waitUpdatedAssessment.ScheduleID]
	userRoomData := make(map[string][]*external.H5PUserContentScore)
	canSetScoreContentMap := make(map[string]*AllowEditScoreContent)
	studentCommentMap := make(map[string]string)
	if hasScore {
		for _, item := range roomData.ScoresByUser {
			if item.User == nil {
				continue
			}
			userRoomData[item.User.UserID] = item.Scores
		}

		canSetScoreContentMap, err = GetAssessmentExternalService().AllowEditScoreContent(ctx, roomData.ScoresByUser)
		if err != nil {
			return err
		}
		studentCommentMap, err = GetAssessmentExternalService().StudentCommentMap(ctx, roomData.TeacherCommentsByStudent)
		if err != nil {
			return err
		}
	}

	if waitUpdatedAssessment.AssessmentType == v2.AssessmentTypeReviewStudy {
		return a.updateReviewStudyAssessment(ctx, op, updateReviewStudyAssessmentInput{
			status:                status,
			req:                   req,
			waitUpdatedAssessment: waitUpdatedAssessment,
			waitUpdatedUsers:      waitUpdatedUsers,
			userIDAndUserTypeMap:  userIDAndUserTypeMap,
			at:                    at,
			userRoomData:          userRoomData,
			canSetScoreContentMap: canSetScoreContentMap,
			studentCommentMap:     studentCommentMap,
		})
	}

	scheduleContents, err := at.FirstGetContentsFromSchedule()
	if err != nil {
		return err
	}

	assessmentContentMap, err := at.FirstGetAssessmentContentMap()
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
	contentOutcomeIDMap := make(map[string][]string, len(scheduleContents))
	for _, item := range scheduleContents {
		contentOutcomeIDMap[item.ID] = item.OutcomeIDs
	}

	outcomeIDs := make([]string, 0)
	for _, item := range scheduleContents {
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

	outcomeFromAssessmentMap, err := at.FirstGetOutcomeFromAssessment()
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
						key := at.GetKey([]string{userItem.ID, contentItem.ID, outcomeID})
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
		assessmentUserItem, ok := userIDAndUserTypeMap[at.GetKey([]string{stuItem.StudentID, v2.AssessmentUserTypeStudent.String()})]
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
					key := at.GetKey([]string{assessmentUserItem.ID, assessmentContentItem.ID, outcomeItem.OutcomeID})
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
				if _, ok := userRoomData[stuItem.StudentID]; ok {
					if canSetScoreContentItem, ok := canSetScoreContentMap[contentItem.ContentID]; ok {
						newScore := &external.H5PSetScoreRequest{
							RoomID:    waitUpdatedAssessment.ScheduleID,
							StudentID: stuItem.StudentID,
							Score:     stuResult.Score,
						}

						newScore.ContentID = canSetScoreContentItem.ContentID
						newScore.SubContentID = canSetScoreContentItem.SubContentID

						newScores = append(newScores, newScore)
					}
				}
			}
		}

		if stuComment, ok := studentCommentMap[stuItem.StudentID]; ok && stuComment != stuItem.ReviewerComment {
			newComment := external.H5PAddRoomCommentRequest{
				RoomID:    waitUpdatedAssessment.ScheduleID,
				StudentID: stuItem.StudentID,
				Comment:   stuItem.ReviewerComment,
			}
			newComments = append(newComments, &newComment)
		} else if stuItem.ReviewerComment != "" {
			newComment := external.H5PAddRoomCommentRequest{
				RoomID:    waitUpdatedAssessment.ScheduleID,
				StudentID: stuItem.StudentID,
				Comment:   stuItem.ReviewerComment,
			}
			newComments = append(newComments, &newComment)
		}
	}

	// update student comment
	err = a.updateStudentCommentAndScore(ctx, op, &updateStudentCommentAndScoreInput{
		assessmentType: waitUpdatedAssessment.AssessmentType,
		scheduleID:     waitUpdatedAssessment.ScheduleID,
		newScores:      newScores,
		newComments:    newComments,
		at:             at,
	})
	if err != nil {
		return err
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

type updateStudentCommentAndScoreInput struct {
	assessmentType v2.AssessmentType
	scheduleID     string
	newScores      []*external.H5PSetScoreRequest
	newComments    []*external.H5PAddRoomCommentRequest
	at             *AssessmentTool
}

func (a *assessmentModelV2) updateStudentCommentAndScore(ctx context.Context, op *entity.Operator, input *updateStudentCommentAndScoreInput) error {
	match := GetAssessmentDetailMatch(input.assessmentType, input.at)
	isAnyoneAttempted, _ := match.MatchAnyOneAttempted()
	if !isAnyoneAttempted {
		return nil
	}
	if len(input.newComments) > 0 {
		if _, err := external.GetH5PRoomCommentServiceProvider().BatchAdd(ctx, op, input.newComments); err != nil {
			log.Warn(ctx, "set student comment error", log.Err(err), log.Any("newComments", input.newComments))
			return err
		}
	}

	// update student score
	if len(input.newScores) > 0 {
		if _, err := external.GetH5PRoomScoreServiceProvider().BatchSet(ctx, op, input.newScores); err != nil {
			log.Warn(ctx, "set student score error", log.Err(err), log.Any("newScores", input.newScores))
			return err
		}
	}

	return nil
}

type updateReviewStudyAssessmentInput struct {
	status                v2.AssessmentStatus
	req                   *v2.AssessmentUpdateReq
	waitUpdatedAssessment *v2.Assessment
	waitUpdatedUsers      []*v2.AssessmentUser
	userIDAndUserTypeMap  map[string]*v2.AssessmentUser
	at                    *AssessmentTool
	userRoomData          map[string][]*external.H5PUserContentScore
	canSetScoreContentMap map[string]*AllowEditScoreContent
	studentCommentMap     map[string]string
}

func (a *assessmentModelV2) updateReviewStudyAssessment(ctx context.Context, op *entity.Operator, input updateReviewStudyAssessmentInput) error {
	match := GetAssessmentDetailMatch(input.waitUpdatedAssessment.AssessmentType, input.at)
	remainingTimeMap, err := match.MatchRemainingTime()
	if err != nil {
		return err
	}
	remainingTime, ok := remainingTimeMap[input.waitUpdatedAssessment.ID]
	if !ok {
		log.Warn(ctx, "not found assessment remaining time", log.Any("waitUpdateAssessment", input.waitUpdatedAssessment))
		return constant.ErrInvalidArgs
	}
	if remainingTime > 0 {
		log.Warn(ctx, "assessment remaining time is greater than 0", log.Int64("remainingTime", remainingTime), log.Any("waitUpdateAssessment", input.waitUpdatedAssessment))
		return constant.ErrInvalidArgs
	}

	// user comment,score
	newScores := make([]*external.H5PSetScoreRequest, 0)
	newComments := make([]*external.H5PAddRoomCommentRequest, 0)

	contentReqMap := make(map[string]*v2.AssessmentUpdateContentReq)
	for _, item := range input.req.Contents {
		contentReqMap[item.ContentID] = item
	}

	for _, stuItem := range input.req.Students {
		if stuItem.Status == v2.AssessmentUserStatusNotParticipate {
			continue
		}
		// verify student data
		_, ok := input.userIDAndUserTypeMap[input.at.GetKey([]string{stuItem.StudentID, v2.AssessmentUserTypeStudent.String()})]
		if !ok {
			log.Warn(ctx, "student not exist", log.Any("userIDAndUserTypeMap", input.userIDAndUserTypeMap), log.Any("stuItem", stuItem))
			return constant.ErrInvalidArgs
		}

		for _, stuResult := range stuItem.Results {
			if contentItem, ok := contentReqMap[stuResult.ContentID]; ok {
				if _, ok := input.userRoomData[stuItem.StudentID]; ok {
					if canSetScoreContentItem, ok := input.canSetScoreContentMap[contentItem.ContentID]; ok {
						newScore := &external.H5PSetScoreRequest{
							RoomID:    input.waitUpdatedAssessment.ScheduleID,
							StudentID: stuItem.StudentID,
							Score:     stuResult.Score,
						}

						newScore.ContentID = canSetScoreContentItem.ContentID
						newScore.SubContentID = canSetScoreContentItem.SubContentID

						newScores = append(newScores, newScore)
					}
				}
			}
		}

		if stuComment, ok := input.studentCommentMap[stuItem.StudentID]; ok && stuComment != stuItem.ReviewerComment {
			newComment := external.H5PAddRoomCommentRequest{
				RoomID:    input.waitUpdatedAssessment.ScheduleID,
				StudentID: stuItem.StudentID,
				Comment:   stuItem.ReviewerComment,
			}
			newComments = append(newComments, &newComment)
		} else if stuItem.ReviewerComment != "" {
			newComment := external.H5PAddRoomCommentRequest{
				RoomID:    input.waitUpdatedAssessment.ScheduleID,
				StudentID: stuItem.StudentID,
				Comment:   stuItem.ReviewerComment,
			}
			newComments = append(newComments, &newComment)
		}
	}

	// update student comment
	err = a.updateStudentCommentAndScore(ctx, op, &updateStudentCommentAndScoreInput{
		assessmentType: v2.AssessmentTypeReviewStudy,
		scheduleID:     input.waitUpdatedAssessment.ScheduleID,
		newScores:      newScores,
		newComments:    newComments,
		at:             input.at,
	})
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	input.waitUpdatedAssessment.UpdateAt = now
	input.waitUpdatedAssessment.Status = input.status
	if input.status == v2.AssessmentStatusComplete {
		input.waitUpdatedAssessment.CompleteAt = now
	}

	err = dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		if _, err := assessmentV2.GetAssessmentDA().UpdateTx(ctx, tx, input.waitUpdatedAssessment); err != nil {
			return err
		}

		if len(input.waitUpdatedUsers) > 0 {
			if _, err := assessmentV2.GetAssessmentUserDA().UpdateTx(ctx, tx, input.waitUpdatedUsers); err != nil {
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
