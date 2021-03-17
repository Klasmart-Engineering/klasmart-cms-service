package model

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
	"time"
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
	condition.Pager = dbo.Pager{Page: 1, PageSize: 1}

	var dataList []*entity.ScheduleFeedback
	err := da.GetScheduleFeedbackDA().Query(ctx, condition, &dataList)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.Any("op", op), log.Any("condition", condition))
		return nil, err
	}
	if len(dataList) <= 0 {
		log.Warn(ctx, "not found", log.Any("op", op), log.Any("condition", condition))
		return nil, constant.ErrRecordNotFound
	}
	feedback := dataList[0]
	result := &entity.ScheduleFeedbackView{
		ScheduleFeedback: entity.ScheduleFeedback{
			ID:         feedback.ID,
			ScheduleID: feedback.ScheduleID,
			UserID:     feedback.UserID,
			Comment:    feedback.Comment,
			CreateAt:   feedback.CreateAt,
		},
	}
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
				AssignmentUrl:  item.Url,
				AssignmentName: item.Name,
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
		//teacherIDs, err := GetScheduleRelationModel().GetTeacherIDs(ctx, op, input.ScheduleID)
		//if err != nil {
		//	return "", err
		//}
		//scheduleInfo, err := GetScheduleModel().GetPlainByID(ctx, input.ScheduleID)
		//if err != nil {
		//	return "", err
		//}
		//classID, err := GetScheduleRelationModel().GetClassRosterID(ctx, op, input.ScheduleID)
		//if err != nil {
		//	return nil, err
		//}
		//homeFun := entity.SaveHomeFunStudyArgs{
		//	ScheduleID:     input.ScheduleID,
		//	ClassID:        classID,
		//	LessonName:     scheduleInfo.Title,
		//	TeacherIDs:     teacherIDs,
		//	StudentID:      op.UserID,
		//	DueAt:          scheduleInfo.DueAt,
		//	LatestSubmitID: feedback.ID,
		//	LatestSubmitAt: feedback.CreateAt,
		//}
		//err = GetHomeFunStudyModel().SaveHomeFunStudy(ctx, op, homeFun)
		//if err != nil {
		//	log.Error(ctx, "insert homeFunStudy error", log.Err(err), log.Any("op", op), log.Any("homeFun", homeFun), log.Any("input", input))
		//	return nil, err
		//}
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
