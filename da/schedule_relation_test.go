package da

import (
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"strings"
	"testing"
)

func TestScheduleFilterSubject(t *testing.T) {
	condition := &ScheduleRelationCondition{
		ScheduleFilterSubject: &ScheduleFilterSubject{
			ProgramID: sql.NullString{
				String: "",
				Valid:  true,
			},
			OrgID: sql.NullString{
				String: "",
				Valid:  true,
			},
			RelationIDs: entity.NullStrings{
				Strings: []string{""},
				Valid:   true,
			},
		},
	}

	wheres, parameters := condition.GetConditions()
	whereSql := strings.Join(wheres, " and ")
	t.Log(whereSql)
	t.Log(parameters)
}

func TestScheduleAndRelationCondition(t *testing.T) {
	condition := &ScheduleRelationCondition{
		ScheduleAndRelations: []*ScheduleAndRelations{
			{
				ScheduleID:  "00000",
				RelationIDs: []string{"11111", "222222", "33333"},
			},
			{
				ScheduleID:  "111111",
				RelationIDs: []string{"3333", "44444", "66666"},
			},
		},
	}
	wheres, parameters := condition.GetConditions()
	whereSql := strings.Join(wheres, " and ")
	t.Log(whereSql)
	t.Log(parameters)
}
