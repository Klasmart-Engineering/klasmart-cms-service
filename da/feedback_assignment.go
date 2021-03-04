package da

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/common-cn/logger"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"sync"
)

type IFeedbackAssignmentDA interface {
	dbo.DataAccesser
	BatchInsert(ctx context.Context, dbContext *dbo.DBContext, assignments []*entity.FeedbackAssignment) (int64, error)
}

type feedbackAssignmentDA struct {
	dbo.BaseDA
}

func (s *feedbackAssignmentDA) BatchInsert(ctx context.Context, dbContext *dbo.DBContext, assignments []*entity.FeedbackAssignment) (int64, error) {
	var data [][]interface{}
	for _, item := range assignments {
		data = append(data, []interface{}{
			item.ID,
			item.FeedbackID,
			item.AssignmentUrl,
			item.AssignmentName,
			item.Number,
			item.CreateAt,
			item.UpdateAt,
			item.DeleteAt,
		})
	}
	sql := SQLBatchInsert(constant.TableNameFeedbackAssignment, []string{
		"`id`",
		"`feedback_id`",
		"`assignment_url`",
		"`assignment_name`",
		"`number`",
		"`create_at`",
		"`update_at`",
		"`delete_at`",
	}, data)
	execResult := dbContext.Exec(sql.Format, sql.Values...)
	if execResult.Error != nil {
		logger.Error(ctx, "db exec sql error", log.Any("sql", sql), log.Err(execResult.Error))
		return 0, execResult.Error
	}
	return execResult.RowsAffected, nil
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
	FeedBackID sql.NullString

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
		return "number asc"
	}
}