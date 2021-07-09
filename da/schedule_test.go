package da

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"testing"

	"gitlab.badanamu.com.cn/calmisland/dbo"

	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

func Test_GetLessonPlanIDsByCondition_Sql(t *testing.T) {
	c := ScheduleCondition{
		Status: sql.NullString{
			String: string(entity.ScheduleStatusClosed),
			Valid:  true,
		},
		// ClassID: sql.NullString{
		// 	String: "Class_1",
		// 	Valid:  true,
		// },
	}
	wheres, parameters := c.GetConditions()
	whereSql := strings.Join(wheres, " and ")
	t.Log(whereSql)
	t.Log(parameters)
}

func initDBForTestSchedule() {
	cfg := &config.Config{
		DBConfig: config.DBConfig{
			ConnectionString: "root:Passw0rd@tcp(127.0.0.1:3306)/kidsloop2?charset=utf8mb4&parseTime=True&loc=Local",
		},
	}
	config.Set(cfg)
	option := dbo.WithConnectionString(cfg.DBConfig.ConnectionString)
	newDBO, err := dbo.NewWithConfig(option)
	if err != nil {
		log.Println("connection mysql db error:", err)
		return
	}
	dbo.ReplaceGlobal(newDBO)
}

func TestGetTeachLoadByCondition(t *testing.T) {
	initDBForTestSchedule()
	ctx := context.Background()
	input := &entity.ScheduleTeachingLoadInput{
		OrgID:      "72e47ef0-92bf-4429-a06f-2014e3d3df4b",
		ClassIDs:   []string{"5751555a-cc18-4662-9ae5-a5ad90569f79", "49b3be6a-d139-4f82-9b77-0acc89525d3f"},
		TeacherIDs: []string{"4fde6e1b-8efe-58e9-a404-51fb98ebf9b8", "42098862-28b1-5417-9800-3b89e557a2b9"},
		TimeRanges: []*entity.ScheduleTimeRange{
			&entity.ScheduleTimeRange{
				StartAt: 1605456000,
				EndAt:   1605544740,
			},
			&entity.ScheduleTimeRange{
				StartAt: 1605456000,
				EndAt:   1605544740,
			},
			&entity.ScheduleTimeRange{
				StartAt: 1605456000,
				EndAt:   1605544740,
			},
		},
	}
	condition := NewScheduleTeachLoadCondition(input)

	GetScheduleDA().GetTeachLoadByCondition(ctx, dbo.MustGetDB(ctx), condition)
}

func TestScheduleDA_SoftDelete(t *testing.T) {
	initDBForTestSchedule()
	ctx := context.Background()
	GetScheduleDA().SoftDelete(ctx, dbo.MustGetDB(ctx), "5fab9ab3aa60e5b2e0e3c023", &entity.Operator{
		UserID: "64a36ec1-7aa2-53ab-bb96-4c4ff752096b",
	})
}
