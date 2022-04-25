package model

import (
	"context"
	"database/sql"
	"errors"
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
	_assessmentFeedbackModel     IAssessmentFeedbackModel
	_assessmentFeedbackModelOnce = sync.Once{}

	ErrOfflineStudyHasCompleted = errors.New("home fun study has completed")
	//ErrOfflineStudyHasNewFeedback = errors.New("home fun study has new feedback")
)

type assessmentFeedbackModel struct {
	AmsServices external.AmsServices
}

type IAssessmentFeedbackModel interface {
	UserSubmitOfflineStudy(ctx context.Context, op *entity.Operator, req *v2.OfflineStudyUserResultAddReq) error

	IsCompleteByScheduleIDs(ctx context.Context, op *entity.Operator, scheduleIDs []string) (map[string]bool, error)
	GetUserResult(ctx context.Context, op *entity.Operator, scheduleIDs []string, userIDs []string) (map[string][]*v2.AssessmentUserResultDBView, error)
}

func (a *assessmentFeedbackModel) GetUserResult(ctx context.Context, op *entity.Operator, scheduleIDs []string, userIDs []string) (map[string][]*v2.AssessmentUserResultDBView, error) {
	_, userResults, err := assessmentV2.GetAssessmentUserResultDA().GetAssessmentUserResultDBView(ctx, &assessmentV2.AssessmentUserResultDBViewCondition{
		OrgID: sql.NullString{
			String: op.OrgID,
			Valid:  true,
		},
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

func (a *assessmentFeedbackModel) IsCompleteByScheduleIDs(ctx context.Context, op *entity.Operator, scheduleIDs []string) (map[string]bool, error) {
	attemptedMap, err := GetAssessmentInternalModel().AnyoneAttemptedByScheduleIDs(ctx, op, scheduleIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(attemptedMap))

	for scheduleID, item := range attemptedMap {
		result[scheduleID] = item.AssessmentStatus == v2.AssessmentStatusComplete
	}

	return result, nil
}

func (a *assessmentFeedbackModel) UserSubmitOfflineStudy(ctx context.Context, op *entity.Operator, req *v2.OfflineStudyUserResultAddReq) error {
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
		log.Warn(ctx, "assessment is completed", log.Any("req", req), log.Any("assessment", assessment))
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

	reviewerFeedbackCond := &assessmentV2.AssessmentUserResultCondition{
		AssessmentUserID: sql.NullString{
			String: assessmentUser.ID,
			Valid:  true,
		},
	}
	var reviewerFeedbacks []*v2.AssessmentReviewerFeedback
	err = assessmentV2.GetAssessmentUserResultDA().Query(ctx, reviewerFeedbackCond, &reviewerFeedbacks)
	if err != nil {
		return err
	}

	if len(reviewerFeedbacks) <= 0 {
		return a.firstSubmit(ctx, op, req, assessment, assessmentUser)
	} else {
		return a.resubmitted(ctx, req, reviewerFeedbacks[0], assessmentUser)
	}
}

func (a *assessmentFeedbackModel) resubmitted(ctx context.Context, req *v2.OfflineStudyUserResultAddReq, reviewerFeedback *v2.AssessmentReviewerFeedback, assessmentUser *v2.AssessmentUser) error {
	now := time.Now().Unix()
	reviewerFeedback.StudentFeedbackID = req.FeedbackID
	reviewerFeedback.UpdateAt = now

	assessmentUser.StatusBySystem = v2.AssessmentUserSystemStatusResubmitted
	assessmentUser.UpdateAt = now

	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		_, err := assessmentV2.GetAssessmentUserResultDA().UpdateTx(ctx, tx, reviewerFeedback)
		if err != nil {
			return err
		}

		_, err = assessmentV2.GetAssessmentUserDA().UpdateTx(ctx, tx, assessmentUser)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Error(ctx, "update data error",
			log.Err(err),
			log.Any("req", req),
			log.Any("reviewerFeedback", reviewerFeedback),
			log.Any("assessmentUser", assessmentUser),
		)
		return err
	}

	return nil
}

func (a *assessmentFeedbackModel) firstSubmit(ctx context.Context, op *entity.Operator, req *v2.OfflineStudyUserResultAddReq, assessment *v2.Assessment, assessmentUser *v2.AssessmentUser) error {
	now := time.Now().Unix()
	userResult := &v2.AssessmentReviewerFeedback{
		ID:                utils.NewID(),
		AssessmentUserID:  assessmentUser.ID,
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

	assessmentUser.StatusBySystem = v2.AssessmentUserSystemStatusDone
	assessmentUser.UpdateAt = now
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
}

func GetAssessmentFeedbackModel() IAssessmentFeedbackModel {
	_assessmentFeedbackModelOnce.Do(func() {
		_assessmentFeedbackModel = &assessmentFeedbackModel{
			AmsServices: external.GetAmsServices(),
		}
	})
	return _assessmentFeedbackModel
}
