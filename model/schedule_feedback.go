package model

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/da"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
)

var (
	ErrOnlyStudentCanSubmitFeedback  = errors.New("only student can submit feedback")
	ErrFeedbackNotGenerateAssessment = errors.New("feedback not generate assessment")
)

type IScheduleFeedbackModel interface {
	Add(ctx context.Context, op *entity.Operator, input *entity.ScheduleFeedbackAddInput) (string, error)
	ExistByScheduleID(ctx context.Context, op *entity.Operator, scheduleID string) (bool, error)
	ExistByScheduleIDs(ctx context.Context, op *entity.Operator, scheduleIDs []string) (map[string]bool, error)
	Query(ctx context.Context, op *entity.Operator, condition *da.ScheduleFeedbackCondition) ([]*entity.ScheduleFeedbackView, error)
	GetNewest(ctx context.Context, op *entity.Operator, userID string, scheduleID string) (*entity.ScheduleFeedbackView, error)
}

type scheduleFeedbackModel struct {
}

func (s *scheduleFeedbackModel) GetNewest(ctx context.Context, op *entity.Operator, userID string, scheduleID string) (*entity.ScheduleFeedbackView, error) {
	result := &entity.ScheduleFeedbackView{
		IsAllowSubmit: true,
	}
	condition := &da.ScheduleFeedbackCondition{
		ScheduleID: sql.NullString{
			String: scheduleID,
			Valid:  true,
		},
		UserID: sql.NullString{
			String: userID,
			Valid:  true,
		},
	}
	// get feedback info
	condition.Pager = dbo.Pager{Page: 1, PageSize: 1}
	var dataList []*entity.ScheduleFeedback
	err := da.GetScheduleFeedbackDA().Query(ctx, condition, &dataList)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.Any("op", op), log.Any("condition", condition))
		return nil, err
	}
	if len(dataList) <= 0 {
		log.Warn(ctx, "not found", log.Any("op", op), log.Any("condition", condition))
		return result, nil
	}
	feedback := dataList[0]
	result.ScheduleFeedback = entity.ScheduleFeedback{
		ID:         feedback.ID,
		ScheduleID: feedback.ScheduleID,
		UserID:     feedback.UserID,
		Comment:    feedback.Comment,
		CreateAt:   feedback.CreateAt,
	}

	// get assignment
	assignmentCondition := &da.FeedbackAssignmentCondition{
		FeedBackID: sql.NullString{
			String: result.ID,
			Valid:  true,
		},
	}
	assignments, err := GetFeedbackAssignmentModel().Query(ctx, op, assignmentCondition)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.Any("op", op), log.Any("assignmentCondition", assignmentCondition))
		return nil, err
	}
	result.Assignments = assignments

	isCompletedScheduleMap, err := GetAssessmentFeedbackModel().IsCompleteByScheduleIDs(ctx, op, []string{scheduleID})
	if err != nil {
		return nil, err
	}
	result.IsAllowSubmit = !isCompletedScheduleMap[scheduleID]

	return result, nil
}

func (s *scheduleFeedbackModel) Query(ctx context.Context, op *entity.Operator, condition *da.ScheduleFeedbackCondition) ([]*entity.ScheduleFeedbackView, error) {
	var dataList []*entity.ScheduleFeedback
	err := da.GetScheduleFeedbackDA().Query(ctx, condition, &dataList)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.Any("op", op), log.Any("condition", condition))
		return nil, err
	}
	if len(dataList) <= 0 {
		log.Info(ctx, "query not found", log.Err(err), log.Any("op", op), log.Any("condition", condition))
		return []*entity.ScheduleFeedbackView{}, nil
	}
	feedbackIDs := make([]string, len(dataList))
	result := make([]*entity.ScheduleFeedbackView, len(dataList))
	for i, item := range dataList {
		result[i] = &entity.ScheduleFeedbackView{ScheduleFeedback: *item}
		feedbackIDs[i] = item.ID
	}
	assignmentCondition := &da.FeedbackAssignmentCondition{
		FeedBackIDs: entity.NullStrings{
			Strings: feedbackIDs,
			Valid:   true,
		},
	}
	assignmentMap, err := GetFeedbackAssignmentModel().QueryMap(ctx, op, assignmentCondition)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.Any("op", op), log.Any("assignmentCondition", assignmentCondition))
		return nil, err
	}
	for _, item := range result {
		item.Assignments = assignmentMap[item.ID]
	}
	return result, nil
}

func (s *scheduleFeedbackModel) ExistByScheduleID(ctx context.Context, op *entity.Operator, scheduleID string) (bool, error) {
	condition := &da.ScheduleFeedbackCondition{
		ScheduleID: sql.NullString{
			String: scheduleID,
			Valid:  true,
		},
	}
	count, err := da.GetScheduleFeedbackDA().Count(ctx, condition, &entity.ScheduleFeedback{})
	if err != nil {
		log.Error(ctx, "da.GetScheduleFeedbackDA().Count error",
			log.Err(err),
			log.Any("op", op),
			log.Any("condition", condition))
		return false, err
	}
	return count > 0, nil
}

func (s *scheduleFeedbackModel) ExistByScheduleIDs(ctx context.Context, op *entity.Operator, scheduleIDs []string) (map[string]bool, error) {
	condition := &da.ScheduleFeedbackCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
	}
	var scheduleFeedbacks []*entity.ScheduleFeedback
	err := da.GetScheduleFeedbackDA().Query(ctx, condition, &scheduleFeedbacks)
	if err != nil {
		log.Error(ctx, "da.GetScheduleFeedbackDA().Query error",
			log.Err(err),
			log.Any("op", op),
			log.Any("condition", condition))
		return nil, err
	}

	result := make(map[string]bool, len(scheduleIDs))
	for _, scheduleFeedback := range scheduleFeedbacks {
		result[scheduleFeedback.ScheduleID] = true
	}

	return result, nil
}

func (s *scheduleFeedbackModel) Add(ctx context.Context, op *entity.Operator, input *entity.ScheduleFeedbackAddInput) (string, error) {
	err := s.verifyScheduleFeedback(ctx, op, input)
	if err != nil {
		return "", err
	}

	// notify classes assignments to statistic home fun attendance
	go func(ctx context.Context, op *entity.Operator) {
		data := &v2.ScheduleEndClassCallBackReq{
			ScheduleID:    input.ScheduleID,
			AttendanceIDs: []string{op.UserID},
			ClassEndAt:    time.Now().Unix(),
		}
		log.Debug(ctx, "feedback notify assignments", log.Any("data", data))
		err := GetClassesAssignmentsModel().CreateRecord(ctx, op, data)
		if err != nil {
			log.Error(ctx, "feedback notify assignments",
				log.Err(err),
				log.Any("data", data))
		}
	}(ctx, op)

	id, err := dbo.GetTransResult(ctx, func(ctx context.Context, tx *dbo.DBContext) (interface{}, error) {
		// insert feedback
		feedback := &entity.ScheduleFeedback{
			ID:         utils.NewID(),
			ScheduleID: input.ScheduleID,
			UserID:     op.UserID,
			Comment:    input.Comment,
			CreateAt:   time.Now().Unix(),
			UpdateAt:   0,
			DeleteAt:   0,
		}
		_, err = da.GetScheduleFeedbackDA().InsertTx(ctx, tx, feedback)
		if err != nil {
			log.Error(ctx, "insert error", log.Err(err), log.Any("op", op), log.Any("input", input))
			return "", err
		}

		// insert feedback assignments
		assignments := make([]*entity.FeedbackAssignment, len(input.Assignments))
		for i, item := range input.Assignments {
			if item.AttachmentID == "" || item.AttachmentName == "" {
				log.Info(ctx, "feedback assignment invalid args", log.Any("input", input))
				return "", constant.ErrInvalidArgs
			}
			assignments[i] = &entity.FeedbackAssignment{
				ID:             utils.NewID(),
				FeedbackID:     feedback.ID,
				AttachmentID:   item.AttachmentID,
				AttachmentName: item.AttachmentName,
				Number:         item.Number,
				CreateAt:       time.Now().Unix(),
				UpdateAt:       0,
				DeleteAt:       0,
			}
		}
		_, err = da.GetFeedbackAssignmentDA().BatchInsert(ctx, tx, assignments)
		if err != nil {
			log.Error(ctx, "feedback assignment insert error", log.Err(err), log.Any("op", op), log.Any("input", input))
			return "", err
		}

		// insert homeFunStudy
		offlineStudyAddReq := &v2.OfflineStudyUserResultAddReq{
			ScheduleID: input.ScheduleID,
			UserID:     op.UserID,
			FeedbackID: feedback.ID,
		}
		err = GetAssessmentFeedbackModel().UserSubmitOfflineStudy(ctx, op, offlineStudyAddReq)
		if err != nil {
			return "", err
		}
		return feedback.ID, nil
	})
	if err != nil {
		log.Error(ctx, "feedback insert error", log.Err(err), log.Any("op", op), log.Any("input", input))
		return "", err
	}
	err = da.GetScheduleRedisDA().Clean(ctx, op.OrgID)
	if err != nil {
		log.Warn(ctx, "clean schedule cache error", log.String("orgID", op.OrgID), log.Err(err))
	}

	go func(ctx context.Context) {
		for _, v := range input.Assignments {
			removeResourceMetadata(ctx, v.AttachmentID)
		}
	}(ctx)

	return id.(string), nil
}

func (s *scheduleFeedbackModel) verifyScheduleFeedback(ctx context.Context, op *entity.Operator, input *entity.ScheduleFeedbackAddInput) error {
	if len(input.Assignments) <= 0 {
		log.Info(ctx, "assignments empty", log.Any("op", op), log.Any("input", input))
		return constant.ErrInvalidArgs
	}
	exist, err := GetScheduleModel().ExistScheduleByID(ctx, input.ScheduleID)
	if err != nil {
		log.Error(ctx, "GetScheduleModel.ExistScheduleByID error", log.Err(err), log.Any("input", input))
		return err
	}
	if !exist {
		log.Info(ctx, "schedule id not found", log.Any("op", op), log.Any("input", input))
		return constant.ErrRecordNotFound
	}
	roleType, err := GetScheduleRelationModel().GetRelationTypeByScheduleID(ctx, op, input.ScheduleID)
	if err != nil {
		log.Error(ctx, "get relation type error", log.Any("op", op), log.Any("input", input), log.Err(err))
		return err
	}
	if roleType != entity.ScheduleRoleTypeStudent {
		log.Info(ctx, "not student", log.String("roleType", string(roleType)), log.Any("op", op), log.Any("input", input))
		return ErrOnlyStudentCanSubmitFeedback
	}
	return nil
}

var (
	_scheduleFeedbackOnce  sync.Once
	_scheduleFeedbackModel IScheduleFeedbackModel
)

func GetScheduleFeedbackModel() IScheduleFeedbackModel {
	_scheduleFeedbackOnce.Do(func() {
		_scheduleFeedbackModel = &scheduleFeedbackModel{}
	})
	return _scheduleFeedbackModel
}
