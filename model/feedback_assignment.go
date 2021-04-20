package model

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IFeedbackAssignmentModel interface {
	Query(ctx context.Context, op *entity.Operator, condition *da.FeedbackAssignmentCondition) ([]*entity.FeedbackAssignmentView, error)
	QueryMap(ctx context.Context, op *entity.Operator, condition *da.FeedbackAssignmentCondition) (map[string][]*entity.FeedbackAssignmentView, error)
}

type feedbackAssignmentModel struct{}

func (f *feedbackAssignmentModel) QueryMap(ctx context.Context, op *entity.Operator, condition *da.FeedbackAssignmentCondition) (map[string][]*entity.FeedbackAssignmentView, error) {
	var assignments []*entity.FeedbackAssignment
	err := da.GetFeedbackAssignmentDA().Query(ctx, condition, &assignments)
	if err != nil {
		log.Error(ctx, "query error", log.Any("op", op), log.Any("condition", condition))
		return nil, err
	}
	result := make(map[string][]*entity.FeedbackAssignmentView)
	for _, assignmentItem := range assignments {
		item := &entity.FeedbackAssignmentView{
			AttachmentID:   assignmentItem.AttachmentID,
			AttachmentName: assignmentItem.AttachmentName,
			Number:         assignmentItem.Number,
		}
		if _, ok := result[assignmentItem.FeedbackID]; ok {
			result[assignmentItem.FeedbackID] = append(result[assignmentItem.FeedbackID], item)
		} else {
			result[assignmentItem.FeedbackID] = []*entity.FeedbackAssignmentView{item}
		}
	}
	return result, nil
}

func (f *feedbackAssignmentModel) Query(ctx context.Context, op *entity.Operator, condition *da.FeedbackAssignmentCondition) ([]*entity.FeedbackAssignmentView, error) {
	var assignments []*entity.FeedbackAssignment
	err := da.GetFeedbackAssignmentDA().Query(ctx, condition, &assignments)
	if err != nil {
		log.Error(ctx, "query error", log.Any("op", op), log.Any("condition", condition))
		return nil, err
	}
	result := make([]*entity.FeedbackAssignmentView, len(assignments))
	for i, item := range assignments {
		result[i] = &entity.FeedbackAssignmentView{
			AttachmentID:   item.AttachmentID,
			AttachmentName: item.AttachmentName,
			Number:         item.Number,
		}
	}
	return result, nil
}

var (
	_feedbackAssignmentOnce  sync.Once
	_feedbackAssignmentModel IFeedbackAssignmentModel
)

func GetFeedbackAssignmentModel() IFeedbackAssignmentModel {
	_feedbackAssignmentOnce.Do(func() {
		_feedbackAssignmentModel = &feedbackAssignmentModel{}
	})
	return _feedbackAssignmentModel
}
