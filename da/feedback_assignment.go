package da

import (
	"context"
	"database/sql"
	"sync"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
)

type IFeedbackAssignmentDA interface {
	dbo.DataAccesser
	BatchInsert(ctx context.Context, dbContext *dbo.DBContext, assignments []*entity.FeedbackAssignment) (int64, error)
}

type feedbackAssignmentDA struct {
	dbo.BaseDA
}

func (s *feedbackAssignmentDA) BatchInsert(ctx context.Context, dbContext *dbo.DBContext, assignments []*entity.FeedbackAssignment) (int64, error) {
	_, err := s.InsertTx(ctx, dbContext, &assignments)
	if err != nil {
		log.Error(ctx, "batch insert feedbackAssignment: batch insert failed", log.Any("assignments", assignments), log.Err(err))
		return 0, err
	}
	return int64(len(assignments)), nil
}

var (
	_feedbackAssignmentOnce sync.Once
	_feedbackAssignmentDA   IFeedbackAssignmentDA
)

func GetFeedbackAssignmentDA() IFeedbackAssignmentDA {
	_feedbackAssignmentOnce.Do(func() {
		_feedbackAssignmentDA = &feedbackAssignmentDA{}
	})
	return _feedbackAssignmentDA
}

type FeedbackAssignmentCondition struct {
	FeedBackID  sql.NullString
	FeedBackIDs entity.NullStrings

	OrderBy FeedbackAssignmentOrderBy
	Pager   dbo.Pager

	DeleteAt sql.NullInt64
}

func (c FeedbackAssignmentCondition) GetConditions() ([]string, []interface{}) {
	var wheres []string
	var params []interface{}

	if c.FeedBackID.Valid {
		wheres = append(wheres, "feedback_id = ?")
		params = append(params, c.FeedBackID.String)
	}

	if c.FeedBackIDs.Valid {
		wheres = append(wheres, "feedback_id in (?)")
		params = append(params, c.FeedBackIDs.Strings)
	}

	if c.DeleteAt.Valid {
		wheres = append(wheres, "delete_at>0")
	} else {
		wheres = append(wheres, "(delete_at=0)")
	}

	return wheres, params
}

func (c FeedbackAssignmentCondition) GetOrderBy() string {
	return c.OrderBy.ToSQL()
}

func (c FeedbackAssignmentCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

type FeedbackAssignmentOrderBy int

func (c FeedbackAssignmentOrderBy) ToSQL() string {
	switch c {
	default:
		return "number asc, attachment_name asc"
	}
}
