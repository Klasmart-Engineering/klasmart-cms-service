package da

import (
	"context"
	"gitlab.badanamu.com.cn/calmisland/dbo"
)

type TeacherLoadLessonsSQLDA struct {
	BaseDA
}

func (c *TeacherLoadLessonsSQLDA) QueryTx(ctx context.Context, tx *dbo.DBContext, condition *TeacherLoadLessonsCondition) (interface{}, error) {
	panic("implement me")
}

type TeacherLoadLessonsCondition struct {
	OrderBy string    `json:"order_by"`
	Pager   dbo.Pager `json:"pager"`
}

func (c *TeacherLoadLessonsCondition) GetConditions() ([]string, []interface{}) {
	wheres := make([]string, 0)
	params := make([]interface{}, 0)

	return wheres, params
}

func (c *TeacherLoadLessonsCondition) GetPager() *dbo.Pager {
	return &c.Pager
}

func (c *TeacherLoadLessonsCondition) GetOrderBy() string {
	return c.OrderBy
}
