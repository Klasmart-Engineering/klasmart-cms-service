package model

import (
	"context"
	"database/sql"
	"errors"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
	"time"
)

var (
	ErrOnlyStudentCanSubmitFeedback  = errors.New("only student can submit feedback")
	ErrFeedbackNotGenerateAssessment = errors.New("feedback not generate assessment")
)

type IScheduleFeedbackModel interface {
	Add(ctx context.Context, op *entity.Operator, input *entity.ScheduleFeedbackAddInput) (string, error)
	ExistByScheduleID(ctx context.Context, op *entity.Operator, scheduleID string) (bool, error)
	Query(ctx context.Context, op *entity.Operator, condition *da.ScheduleFeedbackCondition) ([]*entity.ScheduleFeedbackView, error)
	GetNewest(ctx context.Context, op *entity.Operator, condition *da.ScheduleFeedbackCondition) (*entity.ScheduleFeedbackView, error)
}

type scheduleFeedbackModel struct {
}

func (s *scheduleFeedbackModel) GetNewest(ctx context.Context, op *entity.Operator, condition *da.ScheduleFeedbackCondition) (*entity.ScheduleFeedbackView, error) {
	result := &entity.ScheduleFeedbackView{
		IsAllowSubmit: true,
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

	homeFun, err := GetHomeFunStudyModel().GetByScheduleIDAndStudentID(ctx, op, feedback.ScheduleID, op.UserID)
	if err == constant.ErrRecordNotFound {
		log.Error(ctx, "not found home fun", log.Err(err), log.Any("op", op), log.Any("feedback", feedback))
		return nil, ErrFeedbackNotGenerateAssessment
	}
	if err != nil {
		log.Error(ctx, "get home fun study  error", log.Err(err), log.Any("op", op), log.Any("feedback", feedback))
		return nil, err
	}
	result.IsAllowSubmit = homeFun.Status != entity.AssessmentStatusComplete
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
		log.Error(ctx, "insert error", log.Err(err), log.Any("op", op), log.String("scheduleID", scheduleID))
		return false, err
	}
	return count > 0, nil
}

func (s *scheduleFeedbackModel) Add(ctx context.Context, op *entity.Operator, input *entity.ScheduleFeedbackAddInput) (string, error) {
	err := s.verifyScheduleFeedback(ctx, op, input)
	if err != nil {
		return "", err
	}

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
		teacherIDs, err := GetScheduleRelationModel().GetTeacherIDs(ctx, op, input.ScheduleID)
		if err != nil {
			return "", err
		}
		scheduleInfo, err := GetScheduleModel().GetPlainByID(ctx, input.ScheduleID)
		if err != nil {
			return "", err
		}
		classID, err := GetScheduleRelationModel().GetClassRosterID(ctx, op, input.ScheduleID)
		if err != nil {
			return "", err
		}
		homeFun := entity.SaveHomeFunStudyArgs{
			ScheduleID:       input.ScheduleID,
			ClassID:          classID,
			LessonName:       scheduleInfo.Title,
			TeacherIDs:       teacherIDs,
			StudentID:        op.UserID,
			DueAt:            scheduleInfo.DueAt,
			LatestFeedbackID: feedback.ID,
			LatestFeedbackAt: feedback.CreateAt,
			SubjectID:        scheduleInfo.SubjectID,
		}
		err = GetHomeFunStudyModel().Save(ctx, tx, op, homeFun)
		if err != nil {
			log.Error(ctx, "insert homeFunStudy error", log.Err(err), log.Any("op", op), log.Any("homeFun", homeFun), log.Any("input", input))
			return "", err
		}
		return feedback.ID, nil
	})
	if err != nil {
		log.Error(ctx, "feedback insert error", log.Err(err), log.Any("op", op), log.Any("input", input))
		return "", err
	}
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
