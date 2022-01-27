package model

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"time"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"

	"gitlab.badanamu.com.cn/calmisland/dbo"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da/assessmentV2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity/v2"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

var (
	assessmentOfflineStudy     IAssessmentOfflineStudyModel
	assessmentOfflineStudyOnce = sync.Once{}

	ErrOfflineStudyHasCompleted   = errors.New("home fun study has completed")
	ErrOfflineStudyHasNewFeedback = errors.New("home fun study has new feedback")
)

type assessmentOfflineStudyModel struct {
	AmsServices external.AmsServices
}

type IAssessmentOfflineStudyModel interface {
	UserSubmitOfflineStudy(ctx context.Context, op *entity.Operator, req *v2.OfflineStudyUserResultAddReq) error
	Draft(ctx context.Context, op *entity.Operator, req *v2.OfflineStudyUserResultUpdateReq) error
	Complete(ctx context.Context, op *entity.Operator, req *v2.OfflineStudyUserResultUpdateReq) error

	GetByID(ctx context.Context, op *entity.Operator, id string) (*v2.GetOfflineStudyUserResultDetailReply, error)
	Page(ctx context.Context, op *entity.Operator, req *v2.AssessmentQueryReq) (*v2.OfflineStudyUserPageReply, error)

	IsAnyOneCompleteByScheduleIDs(ctx context.Context, op *entity.Operator, scheduleIDs []string) (map[string]bool, error)
	GetUserResult(ctx context.Context, op *entity.Operator, scheduleIDs []string, userIDs []string) (map[string][]*v2.AssessmentUserResultDBView, error)
}

func (a *assessmentOfflineStudyModel) GetUserResult(ctx context.Context, op *entity.Operator, scheduleIDs []string, userIDs []string) (map[string][]*v2.AssessmentUserResultDBView, error) {
	_, userResults, err := assessmentV2.GetAssessmentUserResultDA().GetAssessmentUserResultDBView(ctx, &assessmentV2.AssessmentUserResultDBViewCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
		UserIDs: entity.NullStrings{
			Strings: userIDs,
			Valid:   true,
		},
	})
	if err != nil {
		return nil, err
	}

	result := make(map[string][]*v2.AssessmentUserResultDBView)
	for _, item := range userResults {
		result[item.ScheduleID] = append(result[item.ScheduleID], item)
	}

	return result, nil
}

func (a *assessmentOfflineStudyModel) IsAnyOneCompleteByScheduleIDs(ctx context.Context, op *entity.Operator, scheduleIDs []string) (map[string]bool, error) {
	_, userResults, err := assessmentV2.GetAssessmentUserResultDA().GetAssessmentUserResultDBView(ctx, &assessmentV2.AssessmentUserResultDBViewCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		}})
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool)
	for _, item := range userResults {
		if item.Status == v2.UserResultProcessStatusComplete {
			result[item.ScheduleID] = true
		}
	}

	return result, nil
}

func (a *assessmentOfflineStudyModel) Draft(ctx context.Context, op *entity.Operator, req *v2.OfflineStudyUserResultUpdateReq) error {
	return a.update(ctx, op, v2.UserResultProcessStatusDraft, req)
}

func (a *assessmentOfflineStudyModel) Complete(ctx context.Context, op *entity.Operator, req *v2.OfflineStudyUserResultUpdateReq) error {
	return a.update(ctx, op, v2.UserResultProcessStatusComplete, req)
}

func (a *assessmentOfflineStudyModel) update(ctx context.Context, op *entity.Operator, status v2.UserResultProcessStatus, req *v2.OfflineStudyUserResultUpdateReq) error {
	userResult := new(v2.AssessmentReviewerFeedback)
	err := assessmentV2.GetAssessmentUserResultDA().Get(ctx, req.ID, userResult)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "get assessment user result by id not found", log.Err(err), log.Any("req", req), log.Any("op", op))
		return constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "get assessment user result by id not found", log.Err(err), log.Any("req", req), log.Any("op", op))
		return err
	}

	if userResult.Status == v2.UserResultProcessStatusComplete {
		log.Warn(ctx, "user result is complete", log.Any("userResult", userResult), log.Any("op", op))
		return ErrOfflineStudyHasCompleted
	}

	now := time.Now().Unix()

	userResult.Status = status
	userResult.AssessScore = req.AssessScore
	userResult.UpdateAt = now
	if status == v2.UserResultProcessStatusComplete {
		userResult.CompleteAt = now
	}
	userResult.StudentFeedbackID = req.AssessFeedbackID
	userResult.ReviewerID = op.UserID
	userResult.ReviewerComment = req.AssessComment

	// assessment user
	assessmentUser := new(v2.AssessmentUser)
	err = assessmentV2.GetAssessmentUserDA().Get(ctx, userResult.AssessmentUserID, assessmentUser)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "get assessment user by id not found", log.Err(err), log.Any("userResult", userResult), log.Any("op", op))
		return constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "get assessment user by id not found", log.Err(err), log.Any("userResult", userResult), log.Any("op", op))
		return err
	}

	// assessment
	assessment := new(v2.Assessment)
	err = assessmentV2.GetAssessmentDA().Get(ctx, assessmentUser.AssessmentID, assessment)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "get assessment by id not found", log.Err(err), log.Any("assessmentUser", assessmentUser), log.Any("op", op))
		return constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "get assessment by id not found", log.Err(err), log.Any("assessmentUser", assessmentUser), log.Any("op", op))
		return err
	}
	newest, err := GetScheduleFeedbackModel().GetNewest(ctx, op, assessmentUser.UserID, assessment.ScheduleID)
	if err != nil {
		return err
	}
	if req.AssessFeedbackID != newest.ScheduleFeedback.ID {
		log.Warn(ctx, "user have new submit", log.Any("newest", newest), log.Any("req", req), log.Any("op", op))
		return ErrOfflineStudyHasNewFeedback
	}

	// outcomes
	userOutcomeIDs := make([]string, len(req.Outcomes))
	userOutcomeReqMap := make(map[string]*v2.OfflineStudyUserOutcomeUpdateReq)
	for i, item := range req.Outcomes {
		userOutcomeIDs[i] = item.OutcomeID
		userOutcomeReqMap[item.OutcomeID] = item
	}
	userOutcomeCond := &assessmentV2.AssessmentUserOutcomeCondition{
		AssessmentUserIDs: entity.NullStrings{
			Strings: []string{userResult.AssessmentUserID},
			Valid:   true,
		},
	}
	var waitUpdateUserOutcomes []*v2.AssessmentUserOutcome
	err = assessmentV2.GetAssessmentUserOutcomeDA().Query(ctx, userOutcomeCond, &waitUpdateUserOutcomes)
	if err != nil {
		log.Warn(ctx, "get assessment user outcomes error", log.Any("userOutcomeCond", userOutcomeCond), log.Any("req", req), log.Any("op", op))
		return err
	}
	for _, item := range waitUpdateUserOutcomes {
		if reqItem, ok := userOutcomeReqMap[item.OutcomeID]; ok {
			if !reqItem.Status.Valid() {
				log.Warn(ctx, "request user outcome status invalid", log.Any("reqItem", reqItem), log.Any("req", req), log.Any("op", op))
				return constant.ErrInvalidArgs
			}
			item.Status = reqItem.Status
			item.UpdateAt = now
		}
	}

	return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		_, err = assessmentV2.GetAssessmentUserResultDA().UpdateTx(ctx, tx, userResult)
		if err != nil {
			log.Error(ctx, "update assessment user result error", log.Err(err), log.Any("userResult", userResult))
			return err
		}

		if len(waitUpdateUserOutcomes) > 0 {
			_, err = assessmentV2.GetAssessmentUserOutcomeDA().UpdateTx(ctx, tx, waitUpdateUserOutcomes)
			if err != nil {
				log.Error(ctx, "update assessment user outcomes error", log.Err(err), log.Any("waitUpdateUserOutcomes", waitUpdateUserOutcomes))
				return err
			}
		}

		return nil
	})
}

func (a *assessmentOfflineStudyModel) GetByID(ctx context.Context, op *entity.Operator, id string) (*v2.GetOfflineStudyUserResultDetailReply, error) {
	userResult := new(v2.AssessmentReviewerFeedback)
	err := assessmentV2.GetAssessmentUserResultDA().Get(ctx, id, userResult)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "get user result not found", log.Err(err), log.String("id", id))
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "get user result error", log.Err(err), log.String("id", id))
		return nil, err
	}

	assessmentUser := new(v2.AssessmentUser)
	err = assessmentV2.GetAssessmentUserDA().Get(ctx, userResult.AssessmentUserID, assessmentUser)
	if err != nil {
		log.Error(ctx, "get assessment user error", log.Err(err), log.Any("userResult", userResult))
		return nil, err
	}

	assessment := new(v2.Assessment)
	err = assessmentV2.GetAssessmentDA().Get(ctx, assessmentUser.AssessmentID, assessment)
	if err != nil {
		log.Error(ctx, "get assessment error", log.Err(err), log.Any("assessmentUser", assessmentUser))
		return nil, err
	}

	schedules, err := GetScheduleModel().GetVariableDataByIDs(ctx, op, []string{assessment.ScheduleID}, &entity.ScheduleInclude{
		ClassRosterClass: true,
	})
	if err != nil {
		return nil, err
	}

	scheduleMap := make(map[string]*entity.ScheduleVariable, len(schedules))
	for _, item := range schedules {
		scheduleMap[item.ID] = item
	}

	schedule, ok := scheduleMap[assessment.ScheduleID]
	if !ok {
		log.Error(ctx, "schedule not found", log.Any("assessment", assessment))
		return nil, constant.ErrRecordNotFound
	}

	result := &v2.GetOfflineStudyUserResultDetailReply{
		ID:            userResult.ID,
		ScheduleID:    schedule.ID,
		Status:        userResult.Status,
		DueAt:         schedule.DueAt,
		CompleteAt:    userResult.CompleteAt,
		FeedbackID:    userResult.StudentFeedbackID,
		AssessScore:   userResult.AssessScore,
		AssessComment: userResult.ReviewerComment,
		Title:         "",
		Teachers:      nil,
		Student:       nil,
		Outcomes:      nil,
	}

	// title
	var className string
	if schedule.ClassRosterClass != nil {
		className = schedule.ClassRosterClass.Name
	}
	titleInput := v2.GenerateAssessmentTitleInput{
		ClassName:    className,
		ScheduleName: schedule.Title,
	}
	title, err := v2.AssessmentTypeOfflineStudy.Title(ctx, titleInput)
	if err != nil {
		return nil, err
	}
	result.Title = title

	// teachers
	teacherCondition := &assessmentV2.AssessmentUserCondition{
		AssessmentID: sql.NullString{
			String: assessment.ID,
			Valid:  true,
		},
		UserType: sql.NullString{
			String: v2.AssessmentUserTypeTeacher.String(),
			Valid:  true,
		},
	}
	var assessmentTeacher []*v2.AssessmentUser
	err = assessmentV2.GetAssessmentUserDA().Query(ctx, teacherCondition, &assessmentTeacher)
	if err != nil {
		log.Error(ctx, "get assessment user by condition error", log.Any("teacherCondition", teacherCondition))
		return nil, err
	}

	userIDs := make([]string, 0, len(assessmentTeacher))
	for _, item := range assessmentTeacher {
		userIDs = append(userIDs, item.UserID)
	}
	userIDs = append(userIDs, assessmentUser.UserID)
	userMap, err := external.GetUserServiceProvider().BatchGetNameMap(ctx, op, userIDs)
	if err != nil {
		log.Error(ctx, "get user info from external error", log.Any("userIDs", userIDs))
		return nil, err
	}

	for _, item := range assessmentTeacher {
		result.Teachers = append(result.Teachers, &entity.IDName{
			ID:   item.UserID,
			Name: userMap[item.UserID],
		})
	}

	// student
	result.Student = &entity.IDName{
		ID:   assessmentUser.UserID,
		Name: userMap[assessmentUser.UserID],
	}

	// user outcome
	userOutcomeCond := assessmentV2.AssessmentUserOutcomeCondition{
		AssessmentUserIDs: entity.NullStrings{
			Strings: []string{userResult.AssessmentUserID},
			Valid:   true,
		},
	}
	var userOutcomes []*v2.AssessmentUserOutcome
	err = assessmentV2.GetAssessmentUserOutcomeDA().Query(ctx, userOutcomeCond, &userOutcomes)
	if err != nil {
		log.Error(ctx, "get user outcomes error", log.Any("userOutcomeCond", userOutcomeCond))
		return nil, err
	}
	outcomeIDs := make([]string, len(userOutcomes))
	userOutcomeMap := make(map[string]*v2.AssessmentUserOutcome)
	for i, item := range userOutcomes {
		outcomeIDs[i] = item.OutcomeID
		userOutcomeMap[item.OutcomeID] = item
	}
	outcomes, err := GetOutcomeModel().GetByIDs(ctx, op, dbo.MustGetDB(ctx), outcomeIDs)
	if err != nil {
		log.Error(ctx, "get outcomes info by ids error", log.Strings("outcomeIDs", outcomeIDs))
		return nil, err
	}
	result.Outcomes = make([]*v2.OfflineStudyUserOutcomeReply, len(outcomes))
	for i, item := range outcomes {
		userOutcomeReplyItem := &v2.OfflineStudyUserOutcomeReply{
			OutcomeID:   item.ID,
			OutcomeName: item.Name,
			Assumed:     item.Assumed,
			Status:      v2.AssessmentUserOutcomeStatusUnknown,
		}

		if userOutcomeItem, ok := userOutcomeMap[item.ID]; ok {
			userOutcomeReplyItem.Status = userOutcomeItem.Status
		} else if item.Assumed {
			userOutcomeReplyItem.Status = v2.AssessmentUserOutcomeStatusAchieved
		}

		result.Outcomes[i] = userOutcomeReplyItem
	}

	return result, nil
}

func (a *assessmentOfflineStudyModel) UserSubmitOfflineStudy(ctx context.Context, op *entity.Operator, req *v2.OfflineStudyUserResultAddReq) error {
	var assessments []*v2.Assessment
	err := assessmentV2.GetAssessmentDA().Query(ctx, &assessmentV2.AssessmentCondition{
		ScheduleID: sql.NullString{
			String: req.ScheduleID,
			Valid:  true,
		},
	}, &assessments)
	if err != nil {
		return err
	}
	if len(assessments) <= 0 {
		return constant.ErrRecordNotFound
	}

	assessment := assessments[0]
	if assessment.Status == v2.AssessmentStatusComplete {
		return ErrOfflineStudyHasCompleted
	}

	var assessmentUsers []*v2.AssessmentUser
	assessmentUserCond := &assessmentV2.AssessmentUserCondition{
		AssessmentID: sql.NullString{
			String: assessment.ID,
			Valid:  true,
		},
		UserIDs: entity.NullStrings{
			Strings: []string{req.UserID},
			Valid:   true,
		},
		UserType: sql.NullString{
			String: v2.AssessmentUserTypeStudent.String(),
			Valid:  true,
		},
	}
	err = assessmentV2.GetAssessmentUserDA().Query(ctx, assessmentUserCond, &assessmentUsers)
	if err != nil {
		return err
	}
	if len(assessmentUsers) <= 0 {
		return constant.ErrRecordNotFound
	}

	assessmentUser := assessmentUsers[0]

	userResultCond := &assessmentV2.AssessmentUserResultCondition{
		AssessmentUserID: sql.NullString{
			String: assessmentUser.ID,
			Valid:  true,
		},
	}
	var userResults []*v2.AssessmentReviewerFeedback
	err = assessmentV2.GetAssessmentUserResultDA().Query(ctx, userResultCond, &userResults)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	if len(userResults) <= 0 {
		userResult := &v2.AssessmentReviewerFeedback{
			ID:                utils.NewID(),
			AssessmentUserID:  assessmentUser.ID,
			Status:            v2.UserResultProcessStatusStarted,
			StudentFeedbackID: req.FeedbackID,
			CreateAt:          now,
		}

		outcomeIDsMap, err := GetScheduleModel().GetLearningOutcomeIDs(ctx, op, []string{req.ScheduleID})
		if err != nil {
			return err
		}
		outcomeIDs := outcomeIDsMap[req.ScheduleID]
		outcomes, err := GetOutcomeModel().GetByIDs(ctx, op, dbo.MustGetDB(ctx), outcomeIDs)

		waitAddUserOutcomes := make([]*v2.AssessmentUserOutcome, len(outcomeIDs))
		for i, item := range outcomes {
			waitAddUserOutcomeItem := &v2.AssessmentUserOutcome{
				ID:                  utils.NewID(),
				AssessmentUserID:    assessmentUser.ID,
				AssessmentContentID: "",
				OutcomeID:           item.ID,
				CreateAt:            now,
				Status:              v2.AssessmentUserOutcomeStatusUnknown,
			}
			if item.Assumed {
				waitAddUserOutcomeItem.Status = v2.AssessmentUserOutcomeStatusAchieved
			}

			waitAddUserOutcomes[i] = waitAddUserOutcomeItem
		}

		assessmentUser.StatusBySystem = v2.AssessmentUserStatusParticipate
		return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
			if assessment.Status == v2.AssessmentStatusNotStarted {
				assessment.Status = v2.AssessmentStatusStarted
				_, err := assessmentV2.GetAssessmentDA().UpdateTx(ctx, tx, assessment)
				if err != nil {
					return err
				}
			}

			_, err := assessmentV2.GetAssessmentUserDA().UpdateTx(ctx, tx, assessmentUser)
			if err != nil {
				return err
			}

			_, err = assessmentV2.GetAssessmentUserResultDA().InsertTx(ctx, tx, userResult)
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
	} else {
		userResult := userResults[0]
		userResult.StudentFeedbackID = req.FeedbackID
		userResult.UpdateAt = time.Now().Unix()
		_, err = assessmentV2.GetAssessmentUserResultDA().Update(ctx, userResult)
		if err != nil {
			return err
		}

		return nil
	}
}

func (a *assessmentOfflineStudyModel) Page(ctx context.Context, op *entity.Operator, req *v2.AssessmentQueryReq) (*v2.OfflineStudyUserPageReply, error) {
	result := &v2.OfflineStudyUserPageReply{}

	permission := new(AssessmentPermission)
	err := permission.SearchAllPermissions(ctx, op)
	if err != nil {
		return nil, err
	}
	queryCondition := &assessmentV2.UserResultPageCondition{
		OrgID:   op.OrgID,
		OrderBy: assessmentV2.NewAssessmentUserResultOrderBy(req.OrderBy),
		Pager: dbo.Pager{
			Page:     req.PageIndex,
			PageSize: req.PageSize,
		},
		TeacherIDs: entity.NullStrings{},
	}

	if permission.OrgPermission.Status.Valid {
		queryCondition.Status = permission.OrgPermission.Status
	} else {
		if permission.SchoolPermission.Status.Valid {
			queryCondition.Status = permission.SchoolPermission.Status
			queryCondition.TeacherIDs.Strings = permission.SchoolPermission.UserIDs
		}
		if permission.MyPermission.Status.Valid {
			queryCondition.TeacherIDs.Strings = append(queryCondition.TeacherIDs.Strings, permission.MyPermission.UserID)
		}
	}

	queryCondition.Status.Strings = strings.Split(req.Status, ",")
	queryCondition.Status.Valid = len(queryCondition.Status.Strings) > 0

	if req.QueryType == v2.QueryTypeTeacherName {
		teachers, err := external.GetTeacherServiceProvider().Query(ctx, op, op.OrgID, req.QueryKey)
		if err != nil {
			return nil, err
		}
		queryCondition.TeacherIDs.Strings = make([]string, len(teachers))
		for i, item := range teachers {
			queryCondition.TeacherIDs.Strings[i] = item.ID
		}
	}
	queryCondition.TeacherIDs.Valid = len(queryCondition.TeacherIDs.Strings) > 0

	log.Debug(ctx, "query condition", log.Any("queryCondition", queryCondition))

	total, userResults, err := assessmentV2.GetAssessmentUserResultDA().PageByCondition(ctx, queryCondition)
	if err != nil {
		return nil, err
	}

	if total <= 0 {
		return result, nil
	}

	scheduleIDs := make([]string, 0, len(userResults))
	userIDs := make([]string, 0, len(userResults))
	assessmentIDs := make([]string, 0, len(userResults))
	dedupMap := make(map[string]struct{})
	for _, item := range userResults {
		if _, ok := dedupMap[item.UserID]; !ok {
			userIDs = append(userIDs, item.UserID)
		}

		if _, ok := dedupMap[item.ScheduleID]; !ok {
			scheduleIDs = append(scheduleIDs, item.ScheduleID)
		}

		if _, ok := dedupMap[item.AssessmentID]; !ok {
			assessmentIDs = append(assessmentIDs, item.AssessmentID)
		}
	}

	schedules, err := GetScheduleModel().GetVariableDataByIDs(ctx, op, scheduleIDs, &entity.ScheduleInclude{
		ClassRosterClass: true,
	})
	if err != nil {
		return nil, err
	}

	scheduleMap := make(map[string]*entity.ScheduleVariable, len(schedules))
	for _, item := range schedules {
		scheduleMap[item.ID] = item
	}

	teacherCondition := &assessmentV2.AssessmentUserCondition{
		AssessmentIDs: entity.NullStrings{
			Strings: assessmentIDs,
			Valid:   true,
		},
		UserType: sql.NullString{
			String: v2.AssessmentUserTypeTeacher.String(),
			Valid:  true,
		},
	}
	var assessmentTeacher []*v2.AssessmentUser
	err = assessmentV2.GetAssessmentUserDA().Query(ctx, teacherCondition, &assessmentTeacher)
	if err != nil {
		log.Error(ctx, "query assessment user error", log.Any("teacherCondition", teacherCondition))
		return nil, err
	}

	assessmentTeacherMap := make(map[string][]string)
	for _, item := range assessmentTeacher {
		if _, ok := dedupMap[item.UserID]; !ok {
			userIDs = append(userIDs, item.UserID)
		}
		assessmentTeacherMap[item.AssessmentID] = append(assessmentTeacherMap[item.AssessmentID], item.UserID)
	}

	userMap, err := external.GetUserServiceProvider().BatchGetNameMap(ctx, op, userIDs)
	if err != nil {
		return nil, err
	}

	resultItems := make([]*v2.OfflineStudyUserPageItem, len(userResults))
	for i, item := range userResults {
		replyItem := &v2.OfflineStudyUserPageItem{
			ID:          item.ID,
			CompleteAt:  item.CompleteAt,
			Status:      item.Status,
			SubmitAt:    item.CreateAt,
			AssessScore: item.AssessScore,
		}
		if scheduleItem, ok := scheduleMap[item.ScheduleID]; ok {
			var className string
			if scheduleItem.ClassRosterClass != nil {
				className = scheduleItem.ClassRosterClass.Name
			}
			titleInput := v2.GenerateAssessmentTitleInput{
				ClassName:    className,
				ScheduleName: scheduleItem.Title,
			}
			title, err := v2.AssessmentTypeOfflineStudy.Title(ctx, titleInput)
			if err != nil {
				return nil, err
			}
			replyItem.Title = title
			replyItem.DueAt = scheduleItem.DueAt
		}

		replyItem.Student = &entity.IDName{
			ID:   item.UserID,
			Name: userMap[item.UserID],
		}

		if teachers, ok := assessmentTeacherMap[item.AssessmentID]; ok {
			for _, teacherID := range teachers {
				replyItem.Teachers = append(replyItem.Teachers, &entity.IDName{
					ID:   teacherID,
					Name: userMap[teacherID],
				})
			}
		}

		resultItems[i] = replyItem
	}
	result.Total = total
	result.Item = resultItems

	return result, nil
}

func GetAssessmentOfflineStudyModel() IAssessmentOfflineStudyModel {
	assessmentOfflineStudyOnce.Do(func() {
		assessmentOfflineStudy = &assessmentOfflineStudyModel{
			AmsServices: external.GetAmsServices(),
		}
	})
	return assessmentOfflineStudy
}
