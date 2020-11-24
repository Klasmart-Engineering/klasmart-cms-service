package da

import (
	"database/sql"
	"strings"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func Test_GetLessonPlanIDsByCondition_Sql(t *testing.T) {
	c := ScheduleCondition{
		Status: sql.NullString{
			String: string(entity.ScheduleStatusClosed),
			Valid:  true,
		},
		ClassID: sql.NullString{
			String: "Class_1",
			Valid:  true,
		},
	}
	wheres, parameters := c.GetConditions()
	whereSql := strings.Join(wheres, " and ")
	t.Log(whereSql)
	t.Log(parameters)
}
