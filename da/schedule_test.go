package da

import (
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"
	"testing"
)

func Test_GetLessonPlanIDsByCondition_Sql(t *testing.T) {
	c := ScheduleCondition{
		TeacherID: sql.NullString{
			String: "Teacher_1",
			Valid:  true,
		},
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
