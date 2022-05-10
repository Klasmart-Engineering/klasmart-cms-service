package da

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
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

func TestGetSubjectIDsByProgramID(t *testing.T) {
	ctx := context.TODO()
	orgID := "f27efd10-000e-4542-bef2-0ccda39b93d3"
	programID := "d1bbdcc5-0d80-46b0-b98e-162e7439058f"
	relationIDs := []string{"1fe2ac5c-bc61-4290-adf5-27eec9e17b5c"}
	subjectIDs, err := GetScheduleRelationDA().GetSubjectIDsByProgramID(ctx, dbo.MustGetDB(ctx), orgID, programID, relationIDs)
	if err != nil {
		t.Fatal()
	}
	t.Log(subjectIDs)
}
